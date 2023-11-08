package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	goaway "github.com/TwiN/go-away"
	"github.com/olahol/melody"
	redis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {

	port := flag.Int("port", 3000, "port to serve on")
	rurl := flag.String("redis", os.Getenv("REDIS_URL"), "redis url")

	flag.Parse()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	opt, e := redis.ParseURL(*rurl)
	if e != nil {
		log.Fatal().Err(e).Msg("redis connect failed " + *rurl)
	}

	ctx := context.Background()
	R := redis.NewClient(opt)

	// retrieve msgs from redis
	if res, e := R.LRange(ctx, "th:chat", 0, -1).Result(); e == nil {
		msgs = make([]*Msg, len(res))
		for i, v := range res {
			ms := strings.Split(v, " ")
			msgs[i] = &Msg{&User{ms[0], ms[1]}, strings.Join(ms[2:], " ")}
		}
	} else {
		log.Error().Err(e).Msg("redis read th:chat error")
	}

	M := melody.New()
	M.Config.PongWait = 25 * time.Second
	M.Config.PingPeriod = M.Config.PongWait * 9 / 10

	// TODO: improve?
	GA := goaway.NewProfanityDetector()

	http.Handle("/", http.FileServer(http.Dir("./build")))

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		M.HandleRequest(w, r)
	})

	M.HandleConnect(func(s *melody.Session) {

		U := NewUser()
		s.Set("user", U)

		usersMu.Lock()
		users[U.ID] = s
		usersMu.Unlock()

		// init weights + relations

		wbotsMu.Lock()
		relsMu.Lock()

		if wbots[U.ID] == nil {
			wbots[U.ID] = make(map[string]int)
		}
		ws := wbots[U.ID]

		if rels[U.ID] == nil {
			rels[U.ID] = make(map[string]float64)
		}
		rel := rels[U.ID]

		id := bots[rand.Intn(len(bots))].USER.ID
		for k := range botmap {
			ws[k] = 1
			rel[k] = float64(rand.Intn(100) - 50)
			if k == id {
				n := rand.Intn(len(bots))
				ws[k] += n
				rel[k] = float64(n * 100 / len(bots))
			}
		}

		wbotsMu.Unlock()
		relsMu.Unlock()

		// send msg history to client
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

		// extract header + body
		hb := strings.Split(string(msg), " ")
		h := hb[0]
		b := hb[1:]

		switch h {

		// user sent msg
		case "m":

			// prevent empty
			// done client-side but also here for good measure
			m := strings.Join(b, " ")
			if strings.TrimSpace(m) == "" {
				return
			}

			msg1 := GA.Censor(removeAccents(m))

			// store + broadcast msg
			m1 := &Msg{U, msg1}
			O := U.MkMsg("m", msg1)
			msgsMu.Lock()
			msgs = append(msgs, m1)
			msgsMu.Unlock()
			M.Broadcast([]byte(O))
			go rwriteMsg(R, ctx, m1)

			// reset bot lim
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

	// separate goroutine for bots
	go func() {

		npcR, e := regexp.Compile(`(?i)#[\dABCDEF]{6}`)
		if e != nil {
			log.Fatal().Err(e).Msg("error compiling npcR")
		}

		for {
			// wait randomly
			// feels more natural
			// also prevents for loop from running too fast
			time.Sleep(time.Duration(float32(rand.Intn(11))/10+.5) * time.Second)

			if botlim <= 0 || len(msgs) == 0 || M.Len() == 0 {
				continue
			}

			// choose bot (weighted)
			lms := msgs[len(msgs)-min(len(msgs), conf.WLastN):]
			rs, lU := calcWs(lms)
			bot := rs[rand.Intn(len(rs))]
			if bot.USER.ID == msgs[len(msgs)-1].USER.ID {
				continue
			}

			// API req/res
			msg, e := reqRes(M, bot, relStr(lms, bot.USER.ID))
			M.Broadcast([]byte(bot.USER.MkMsg("-t", "")))
			if e != nil {
				log.Error().Err(e).Msg("post error")
				continue
			}

			// store + broadcast msg
			m := &Msg{bot.USER, msg}
			O := bot.USER.MkMsg("m", msg)
			msgsMu.Lock()
			msgs = append(msgs, m)
			msgsMu.Unlock()
			M.Broadcast([]byte(O))
			go rwriteMsg(R, ctx, m)

			// check for user mentions
			// else get last user msg
			id := npcR.FindString(msg)
			if id == "" && strings.Contains(strings.ToLower(msg), "mortal") {
				id = lU
			} else {
				id = "NPC" + id
			}

			// update relation based on sentiment
			if id != "" {
				if rs, ok := rels[id]; ok {
					r := sent.PolarityScores(msg).Compound
					rs[bot.USER.ID] += max(-100, min(100, conf.MaxR*r))
					if r != 0 {
						go func() {
							users[id].Write([]byte(bot.USER.MkMsg("r", fmt.Sprint(r))))
						}()
					}
				}
			}

			botlimMu.Lock()
			botlim--
			botlimMu.Unlock()
		}
	}()

	log.Info().Msgf("Listening on port %d", *port)
	log.Fatal().Err(http.ListenAndServe(fmt.Sprint(":", *port), nil)).Msg("server error")
}

func rwriteMsg(R *redis.Client, ctx context.Context, m *Msg) {
	if e := R.RPush(ctx, "th:chat", m.USER.String()+" "+m.BODY).Err(); e != nil {
		log.Error().Err(e).Msg("redis write th:chat error")
	}
}
