package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/molarmanful/talking-heads/user"
	"github.com/rs/zerolog/log"
)

type Post struct {
	Output []string `json:"output"`
}

func post(msgs []*Msg, sys string) (string, error) {

	j, e := json.Marshal(map[string]interface{}{
		"version": "02e509c789964a7ea8736978a43525956ef40397be9033abf9fd2badfe68c9e3",
		"input": map[string]interface{}{
			"debug":          false,
			"top_k":          50,
			"top_p":          1,
			"prompt":         rep(msgs, 5),
			"temperature":    1,
			"system_prompt":  sys,
			"max_new_tokens": 64,
			"min_new_tokens": -1,
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

	log.Info().Msg("thinking...")
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
	e = json.NewDecoder(res.Body).Decode(post)
	if e != nil {
		return "", e
	}

	O := strings.TrimSpace(strings.Join(post.Output, ""))
	// TODO: remove this
	log.Info().Msg(O)
	return O, nil
}

type Msg struct {
	ID   string
	BODY string
}

func rep(msgs []*Msg, n int) string {

	n = min(len(msgs), n)
	O := make([]string, n)
	for i, m := range msgs[len(msgs)-n:] {
		O[i] = m.Wrap(m.ID)
	}

	return strings.Join(O, "\n")
}

func (m *Msg) Wrap(id string) string {

	if m.ID == id {
		return m.BODY
	}

	return "[INST] " + m.ID + ": " + m.BODY + " [/INST]"
}

type Bot struct {
	USER   *user.User
	PROMPT string
}
