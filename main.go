package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	nanoid "github.com/matoous/go-nanoid/v2"
	"github.com/olahol/melody"
)

func main() {
	m := melody.New()

	http.Handle("/", http.FileServer(http.Dir("./build")))

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		m.HandleRequest(w, r)
	})

	m.HandleConnect(func(s *melody.Session) {
		id, err := nanoid.New()
		if err != nil {
			panic(err)
		}

		s.Set("user", &User{id, "000000"})
		m.Broadcast([]byte("+ " + id))

		log.Println("+", id)
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		v, x := s.Get("user")
		if !x {
			return
		}

		user := v.(*User)
		m.Broadcast(append([]byte("m "+user.String()+" "), msg...))
	})

	m.HandleDisconnect(func(s *melody.Session) {
		v, x := s.Get("user")
		if !x {
			return
		}

		user := v.(*User)
		m.BroadcastOthers([]byte("- "+user.ID), s)

		log.Println("-", user.ID)
	})

	port := flag.Int("port", 3000, "port to serve on")
	log.Println("Listening on port", *port)

	err := http.ListenAndServe(fmt.Sprint(":", *port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

// TODO: random color
type User struct {
	ID, COLOR string
}

func (u *User) String() string {
	return u.ID + " " + u.COLOR
}
