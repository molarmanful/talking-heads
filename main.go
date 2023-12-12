package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {

	port := flag.Int("port", 3000, "port to serve on")
	rurl := flag.String("redis", os.Getenv("REDIS_URL"), "redis url")

	flag.Parse()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	st := New()
	st.InitR(*rurl)

	v, e := st.GetMsgs()
	if e != nil {
		log.Fatal().Err(e).Msg("redis read th:chat error")
	}
	st.Msgs = v

	http.Handle("/", http.FileServer(http.Dir("./build")))

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		st.M.HandleRequest(w, r)
	})

	st.M.HandleConnect(st.WSConn)
	st.M.HandleMessage(st.WSMsg)
	st.M.HandleDisconnect(st.WSDisconn)
	go st.BotLoop()
	go st.NewBotLoop()

	log.Info().Msgf("Listening on port %d", *port)
	log.Fatal().Err(http.ListenAndServe(fmt.Sprint(":", *port), nil)).Msg("server error")
}
