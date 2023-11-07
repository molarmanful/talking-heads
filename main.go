package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	goaway "github.com/TwiN/go-away"
	"github.com/cdipaolo/sentiment"
	"github.com/olahol/melody"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	M := melody.New()
	M.Config.PongWait = 25 * time.Second
	M.Config.PingPeriod = M.Config.PongWait * 9 / 10
	GA := goaway.NewProfanityDetector()

	http.Handle("/", http.FileServer(http.Dir("./build")))

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		M.HandleRequest(w, r)
	})

	M.HandleConnect(func(s *melody.Session) {

		U := NewUser()
		s.Set("user", U)

		wbotsMu.Lock()
		relsMu.Lock()

		if wbots[U.ID] == nil {
			wbots[U.ID] = make(map[string]int)
		}
		ws := wbots[U.ID]

		if rels[U.ID] == nil {
			rels[U.ID] = make(map[string]int)
		}
		rel := rels[U.ID]

		id := bots[rand.Intn(len(bots))].USER.ID
		for k := range botmap {
			ws[k] = 1
			rel[k] = 0
			if k == id {
				n := rand.Intn(len(bots))
				ws[k] += n
				rel[k] = n * 100 / len(bots)
			}
		}

		wbotsMu.Unlock()
		relsMu.Unlock()

		usersMu.Lock()
		users[U.ID] = s
		usersMu.Unlock()

		ms := make([]string, len(msgs))
		for i, v := range msgs {
			ms[i] = v.String()
		}
		s.Write([]byte(U.MkMsg("w", id+"\n"+strings.Join(ms, "\n"))))

		O := U.MkMsg("+", "")
		M.Broadcast([]byte(O))
		log.Info().Msg(O)
	})

	M.HandleMessage(func(s *melody.Session, msg []byte) {

		v, x := s.Get("user")
		if !x {
			return
		}
		U := v.(*User)

		hb := strings.Split(string(msg), " ")
		h := hb[0]
		b := hb[1:]

		switch h {

		case "m":
			m := strings.Join(b, " ")
			if strings.TrimSpace(m) == "" {
				return
			}

			msg1 := GA.Censor(removeAccents(m))

			O := U.MkMsg("m", msg1)
			msgsMu.Lock()
			msgs = append(msgs, &Msg{U, msg1})
			msgsMu.Unlock()
			M.Broadcast([]byte(O))

			botlimMu.Lock()
			botlim = botlimw[rand.Intn(len(botlimw))]
			botlimMu.Unlock()
		}
	})

	M.HandleDisconnect(func(s *melody.Session) {

		v, x := s.Get("user")
		if !x {
			return
		}
		U := v.(*User)

		delete(wbots, U.ID)

		O := U.MkMsg("-", "")
		M.BroadcastOthers([]byte(O), s)

		log.Info().Msg(O)
	})

	go func() {

		npcR, e := regexp.Compile(`(?i)#[\dABCDEF]{6}`)
		if e != nil {
			log.Fatal().Err(e).Msg("error compiling npcR")
		}

		for {

			time.Sleep(time.Duration(float32(rand.Intn(11))/10+.5) * time.Second)

			t0 := time.Now()
			if botlim <= 0 || len(msgs) == 0 || M.Len() == 0 {
				continue
			}

			lms := msgs[len(msgs)-min(len(msgs), 10):]

			rs, lU := calcWs(lms)
			bot := rs[rand.Intn(len(rs))]
			if bot.USER.ID == msgs[len(msgs)-1].USER.ID {
				continue
			}

			msg, e := postReq(M, bot, relStr(lms, bot.USER.ID))
			if e != nil {
				log.Error().Err(e).Msg("post error")
				continue
			}

			M.Broadcast([]byte(bot.USER.MkMsg("-t", "")))

			O := bot.USER.MkMsg("m", msg)
			msgsMu.Lock()
			msgs = append(msgs, &Msg{bot.USER, msg})
			msgsMu.Unlock()
			M.Broadcast([]byte(O))

			id := npcR.FindString(msg)
			if id == "" {
				id = lU
			} else {
				id = "NPC#" + id
			}

			if id != "" {
				if rs, ok := rels[id]; ok {
					s := "gained"
					if r := rs[bot.USER.ID]; sent.SentimentAnalysis(msg, sentiment.English).Score > 0 {
						r = min(100, r+rand.Intn(20)+1)
					} else {
						r = max(-100, r-rand.Intn(20)+1)
						s = "lost"
					}
					users[id].Write([]byte(bot.USER.MkMsg("r", s)))
				}
			}

			botlimMu.Lock()
			botlim--
			botlimMu.Unlock()
			time.Sleep(max(0, 10*time.Second-time.Now().Sub(t0)))
		}
	}()

	port := flag.Int("port", 3000, "port to serve on")
	log.Info().Msgf("Listening on port %d", *port)
	log.Fatal().Err(http.ListenAndServe(fmt.Sprint(":", *port), nil)).Msg("server error")
}
