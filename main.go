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

	c := New()
	c.InitR(*rurl)

	v, e := c.GetMsgs()
	if e != nil {
		log.Fatal().Err(e).Msg("redis read th:chat error")
	}
	c.Msgs = v

	http.Handle("/", http.FileServer(http.Dir("./build")))

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		c.M.HandleRequest(w, r)
	})

	c.M.HandleConnect(c.WSConn)
	c.M.HandleMessage(c.WSMsg)
	c.M.HandleDisconnect(c.WSDisconn)
	go c.BotLoop()

	log.Info().Msgf("Listening on port %d", *port)
	log.Fatal().Err(http.ListenAndServe(fmt.Sprint(":", *port), nil)).Msg("server error")
}
