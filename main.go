package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	goaway "github.com/TwiN/go-away"
	"github.com/molarmanful/talking-heads/user"
	"github.com/olahol/melody"
)

func main() {
	m := melody.New()
	ga := goaway.NewProfanityDetector()

	http.Handle("/", http.FileServer(http.Dir("./build")))

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		m.HandleRequest(w, r)
	})

	m.HandleConnect(func(s *melody.Session) {
		u := user.New()

		s.Set("user", u)
		m.Broadcast([]byte("+ " + u.ID))

		log.Println("+", u.ID)
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		v, x := s.Get("user")
		if !x {
			return
		}

		msg1, e := removeAccents(string(msg[:]))
		if e != nil {
			return
		}
		msg1 = ga.Censor(msg1)

		u := v.(*user.User)
		m.Broadcast([]byte("m " + u.String() + " " + msg1))
	})

	m.HandleDisconnect(func(s *melody.Session) {
		v, x := s.Get("user")
		if !x {
			return
		}

		u := v.(*user.User)
		m.BroadcastOthers([]byte("- "+(*u).ID), s)

		log.Println("-", u.ID)
	})

	port := flag.Int("port", 3000, "port to serve on")
	log.Println("Listening on port", *port)

	e := http.ListenAndServe(fmt.Sprint(":", *port), nil)
	if e != nil {
		log.Fatal(e)
	}
}

func removeAccents(s string) (string, error) {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	o, _, e := transform.String(t, s)
	return o, e
}
