package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"

	goaway "github.com/TwiN/go-away"
	"github.com/olahol/melody"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
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
		ws := wbots[U.ID]
		for i := range ws {
			n := rand.Intn(10) + 1
			for j := 0; j < n*n; j++ {
				ws = append(ws, bots[i])
			}
		}
		wbotsMu.Unlock()

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

			msg1, e := removeAccents(m)
			if e != nil {
				s.Write([]byte(U.MkMsg("e", "send failed")))
				log.Error().Err(e).Msg("error in removeAccents")
				return
			}
			msg1 = GA.Censor(msg1)

			O := U.MkMsg("m", msg1)
			msgsMu.Lock()
			msgs = append(msgs, &Msg{U.ID, msg1})
			msgsMu.Unlock()
			M.Broadcast([]byte(O))

			botlimMu.Lock()
			botlim = rand.Intn(maxbotlim)
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
		for {

			time.Sleep(time.Duration(float32(rand.Intn(11))/10+.5) * time.Second)

			t0 := time.Now()
			if botlim <= 0 || len(msgs) == 0 || M.Len() == 0 {
				continue
			}

			lID := msgs[len(msgs)-1].ID
			rs := bots
			if v, ok := wbots[lID]; ok {
				rs = v
			}

			bot := rs[rand.Intn(len(rs))]
			if bot.USER.ID == lID {
				continue
			}

			msg, e := post(bot.USER.ID, bot.PROMPT)
			if e != nil {
				log.Error().Err(e).Msg("post error")
				continue
			}

			O := bot.USER.MkMsg("m", msg)
			msgsMu.Lock()
			msgs = append(msgs, &Msg{bot.USER.ID, msg})
			msgsMu.Unlock()
			M.Broadcast([]byte(O))

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

func removeAccents(s string) (string, error) {

	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	o, _, e := transform.String(t, s)

	return o, e
}
