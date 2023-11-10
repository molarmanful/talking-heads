package main

import (
	nanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rs/zerolog/log"
)

type User struct {
	ID    string
	COLOR string
}

// Gens new user.
func NewUser() *User {

	id, e := nanoid.Generate("0123456789ABCDEF", 6)
	if e != nil {
		log.Fatal().Err(e).Send()
	}
	return &User{"NPC#" + id, "#" + id}
}

// Gets user as string.
func (u *User) String() string {
	return u.ID + " " + u.COLOR
}

// Gens msg from header + user + body.
func (u *User) MkMsg(h string, b string) string {
	return h + " " + u.String() + " " + b
}
