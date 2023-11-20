package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/olahol/melody"
	"github.com/rs/zerolog/log"
	"go.uber.org/ratelimit"
)

func (ST *State) WSConn(s *melody.Session) {

	U := NewUser()
	s.Set("user", U)

	s.Set("rl", ratelimit.New(5))

	ST.UsersMu.Lock()
	ST.Users[U.ID] = s
	ST.UsersMu.Unlock()

	// init weights + relations

	ST.WbotsMu.Lock()
	ST.RelsMu.Lock()

	if ST.Wbots[U.ID] == nil {
		ST.Wbots[U.ID] = make(map[string]int)
	}
	ws := ST.Wbots[U.ID]

	if ST.Rels[U.ID] == nil {
		ST.Rels[U.ID] = make(map[string]float64)
	}
	rel := ST.Rels[U.ID]

	id := ST.Bots[rand.Intn(len(ST.Bots))].USER.ID
	for k := range ST.BotMap {
		ws[k] = 1
		rel[k] = float64(rand.Intn(100) - 50)
		if k == id {
			n := rand.Intn(len(ST.Bots))
			ws[k] += n
			rel[k] = float64(n * 100 / len(ST.Bots))
		}
	}

	ST.WbotsMu.Unlock()
	ST.RelsMu.Unlock()

	ST.M.BroadcastOthers([]byte(U.MkMsg("+", "")), s)

	// send msg history to client

	lms := ST.Msgs[len(ST.Msgs)-min(len(ST.Msgs), ST.MsgCh):]
	ms := make([]string, len(lms))
	for i, v := range lms {
		ms[i] = v.String()
	}

	s.Write([]byte(U.MkMsg("w", id+"\n"+strings.Join(ms, "\n"))))
	s.Set("chn", 1)

	// send user list to client

	bs := make([]string, len(ST.Bots))
	for i, b := range ST.Bots {
		bs[i] = b.USER.String()
	}

	us := make([]string, len(ST.Users))
	i := 0
	for _, s1 := range ST.Users {
		v, x := s1.Get("user")
		if !x {
			continue
		}
		us[i] = v.(*User).String()
		i++
	}

	s.Write([]byte(U.MkMsg("u", strings.Join(append(bs, us...), "\n"))))
}

func (ST *State) WSMsg(s *melody.Session, msg []byte) {

	// rate-limit
	v, x := s.Get("rl")
	if !x {
		log.Warn().Msg("rl not found")
		return
	}
	rl := v.(ratelimit.Limiter)
	rl.Take()

	v, x = s.Get("user")
	if !x {
		return
	}
	U := v.(*User)

	// extract header + body
	hb := strings.SplitN(string(msg), " ", 2)
	println(string(msg))
	h := hb[0]

	switch h {

	// user sent msg
	case "m":
		// prevent empty
		// done client-side but also here for good measure
		b := hb[1]
		if strings.TrimSpace(b) == "" {
			return
		}

		// store + broadcast msg
		m := &Msg{*U, b}
		ST.MsgsMu.Lock()
		ST.Msgs = append(ST.Msgs, m)
		ST.MsgsMu.Unlock()
		ST.M.Broadcast([]byte("m " + m.String()))
		go ST.StoreMsg(m)

		// reset bot lim
		ST.BotLimMu.Lock()
		ST.BotLim = ST.BotLimW[rand.Intn(len(ST.BotLimW))]
		ST.BotLimMu.Unlock()

	// user needs more messages
	case "g":
		v, x := s.Get("chn")
		if !x {
			return
		}

		// locks exist client-side but also here for good measure
		w, x := s.Get("chnL")
		if x && w.(bool) {
			return
		}
		s.Set("chnL", true)

		chn := v.(int)
		lms := ST.Msgs[len(ST.Msgs)-min(len(ST.Msgs), (chn+1)*ST.MsgCh) : len(ST.Msgs)-min(len(ST.Msgs), chn*ST.MsgCh)]
		if len(lms) == 0 {
			return
		}

		ms := make([]string, len(lms))
		for i, v := range lms {
			ms[i] = v.String()
		}

		s.Write([]byte(U.MkMsg("g", strings.Join(ms, "\n"))))
		chn++
		s.Set("chn", chn)
		s.Set("chnL", false)
	}
}

func (ST *State) BotLoop() {

	npcR := regexp.MustCompile(`(?i)#[\dABCDEF]{6}`)

	for {
		time.Sleep(0)

		if ST.BotLim <= 0 || len(ST.Msgs) == 0 || ST.M.Len() == 0 {
			continue
		}

		// choose bot (weighted)
		lms := ST.Msgs[len(ST.Msgs)-min(len(ST.Msgs), ST.WLastN):]
		rs, lU := ST.CalcWs(lms)
		bot := rs[rand.Intn(len(rs))]
		if bot.USER.ID == ST.Msgs[len(ST.Msgs)-1].USER.ID {
			ST.BotLim = 0
			continue
		}

		// wait randomly
		// feels more natural
		// also prevents for loop from running too fast
		time.Sleep(time.Duration(float32(rand.Intn(11))/10+.5) * time.Second)

		// API req/res
		msg, e := ST.ReqResGod(bot, ST.RelStr(lms, bot.USER.ID))
		ST.M.Broadcast([]byte(bot.USER.MkMsg("-t", "")))
		if e != nil {
			log.Error().Err(e).Msg("post error")
			continue
		}

		// store + broadcast msg
		m := &Msg{*bot.USER, msg}
		ST.MsgsMu.Lock()
		ST.Msgs = append(ST.Msgs, m)
		ST.MsgsMu.Unlock()
		ST.M.Broadcast([]byte("m " + m.String()))
		go ST.StoreMsg(m)

		// check for user mentions
		// else get last user msg
		id := npcR.FindString(msg)
		if id != "" {
			id = "NPC" + id
		} else {
			id = lU
		}

		// update relation based on sentiment
		if id != "" {
			if rs, ok := ST.Rels[id]; ok {
				go func() {

					n, e := ST.ReqResFeels(msg)
					if e != nil {
						log.Error().Err(e).Msg("sentiment error")
						return
					}
					if -.05 <= n && n <= .05 {
						return
					}

					ST.RelsMu.Lock()
					rs[bot.USER.ID] = max(-100, min(100, rs[bot.USER.ID]+ST.MaxR*n))
					ST.RelsMu.Unlock()
					ST.Users[id].Write([]byte(bot.USER.MkMsg("r", fmt.Sprint(n))))
				}()
			}
		}

		ST.BotLimMu.Lock()
		ST.BotLim--
		ST.BotLimMu.Unlock()
	}
}

func (ST *State) WSDisconn(s *melody.Session) {

	v, x := s.Get("user")
	if !x {
		return
	}
	U := v.(*User)

	ST.WbotsMu.Lock()
	delete(ST.Wbots, U.ID)
	ST.WbotsMu.Unlock()

	O := U.MkMsg("-", "")
	ST.M.BroadcastOthers([]byte(O), s)
}
