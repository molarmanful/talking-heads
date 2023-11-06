package user

import (
	nanoid "github.com/matoous/go-nanoid/v2"
)

type User struct {
	ID    string
	COLOR string
}

// TODO: random color
func New() *User {

	id, e := nanoid.New(4)
	if e != nil {
		panic(e)
	}

	c, e := nanoid.Generate("0123456789", 3)
	if e != nil {
		panic(e)
	}

	return &User{"NPC#" + id, "#" + c}
}

func (u *User) String() string {
	return u.ID + " " + u.COLOR
}

func (u *User) MkMsg(h string, b string) string {
	return h + " " + u.String() + " " + b
}
