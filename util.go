package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	nanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rs/zerolog/log"
)

func post(id string, sys string) (string, error) {

	log.Info().Msg("Q: " + id)

	r := rep(id, 7)
	j, e := json.Marshal(map[string]interface{}{
		// "version": "02e509c789964a7ea8736978a43525956ef40397be9033abf9fd2badfe68c9e3", // llama 2 70b
		"version": "f4e2de70d66816a838a89eeeb621910adffb0dd0baba3976c96980970978018d", // llama 2 13b
		"input": map[string]interface{}{
			"top_k":              50,
			"top_p":              .95,
			"prompt":             r,
			"temperature":        .8,
			"system_prompt":      sys + `Concise one-sentence response to the conversation as ` + id + `, without using speaker labels and ensuring relevance to the context provided. Start your response with "A:". Example:\nA: Witness my power mere mortal!`,
			"max_new_tokens":     64,
			"min_new_tokens":     -1,
			"repetition_penalty": 1.18,
		},
	})
	if e != nil {
		return "", e
	}

	req, e := http.NewRequest("POST", "https://replicate-api-proxy.glitch.me/create_n_get", bytes.NewBuffer(j))
	if e != nil {
		return "", e
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, e := client.Do(req)
	if e != nil {
		return "", e
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", errors.New(res.Status)
	}

	post := &Post{}
	if e = json.NewDecoder(res.Body).Decode(post); e != nil {
		return "", e
	}

	O := strings.TrimSpace(strings.Join(post.Output, ""))
	log.Info().Msg(O)
	return strings.TrimSpace(strings.TrimPrefix(O, "A:")), nil
}

func rep(id string, n int) string {

	n = min(len(msgs), n)
	O := ""
	ins := false
	for _, m := range msgs[len(msgs)-n:] {
		if m.ID == id {
			if ins {
				O = O[:len(O)-1] + " [/INS]\n"
			}
			O += m.BODY + "\n"
			ins = false
			break
		}

		if !ins {
			O += "[INS] "
		}
		O += m.ID + ": " + m.BODY + "\n"
		ins = true
	}
	if ins {
		O = O[:len(O)-1] + " [/INS]\n"
	}

	return O
}

type Post struct {
	Output []string `json:"output"`
}

type Msg struct {
	ID   string
	BODY string
}

type Bot struct {
	USER   *User
	PROMPT string
}

type WPair struct {
	Weights []int
	Sum     int
}

type User struct {
	ID    string
	COLOR string
}

func NewUser() *User {

	id, e := nanoid.Generate("0123456789ABCDEF", 6)
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
