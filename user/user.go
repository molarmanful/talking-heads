package user

import (
	nanoid "github.com/matoous/go-nanoid/v2"
)

type User struct {
	ID    string
	COLOR string
}

func New() *User {

	id, e := nanoid.Generate("0123456789", 6)
	if e != nil {
		panic(e)
	}
	return &User{"NPC#" + id, "#" + id}
}

func (u *User) String() string {
	return u.ID + " " + u.COLOR
}

func (u *User) MkMsg(h string, b string) string {
	return h + " " + u.String() + " " + b
}
