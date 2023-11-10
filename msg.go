package main

import "strings"

type Msg struct {
	USER User
	BODY string
}

// Gets msg as string.
func (m *Msg) String() string {
	return m.USER.String() + " " + m.BODY
}

func ToMsg(s string) *Msg {
	m := strings.SplitN(s, " ", 3)
	return &Msg{User{m[0], m[1]}, m[2]}
}
