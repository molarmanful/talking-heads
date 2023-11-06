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

func post(msgs []*Msg, id string, sys string) (string, error) {

	r := rep(msgs, id, 7)
	log.Info().Msg(r)
	j, e := json.Marshal(map[string]interface{}{
		// "version": "02e509c789964a7ea8736978a43525956ef40397be9033abf9fd2badfe68c9e3", // llama 2 70b
		// "version": "f4e2de70d66816a838a89eeeb621910adffb0dd0baba3976c96980970978018d", // llama 2 13b
		// "version": "18f253bfce9f33fe67ba4f659232c509fbdfb5025e5dbe6027f72eeb91c8624b", // llama 2 13b Q
		// "input": map[string]interface{}{
		// 	"top_k":              50,
		// 	"top_p":              .95,
		// 	"prompt":             r,
		// 	"temperature":        .8,
		// 	"system_prompt":      sys,
		// 	"max_new_tokens":     64,
		// 	"min_new_tokens":     -1,
		// 	"repetition_penalty": 1.18,
		// },
		"version": "6282abe6a492de4145d7bb601023762212f9ddbbe78278bd6771c8b3b2f2a13b", // vicuna 13b
		"input": map[string]interface{}{
			"top_p":              .95,
			"prompt":             r + "\n\n" + sys,
			"temperature":        .8,
			"max_length":         64,
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
	e = json.NewDecoder(res.Body).Decode(post)
	if e != nil {
		return "", e
	}

	O := strings.TrimPrefix(strings.TrimSpace(strings.Join(post.Output, "")), id+": ")
	// TODO: remove this
	log.Info().Msg(O)
	return O, nil
}

type Msg struct {
	ID   string
	BODY string
}

func rep(msgs []*Msg, id string, n int) string {

	n = min(len(msgs), n)
	O := make([]string, n)
	for i, m := range msgs[len(msgs)-n:] {
		O[i] = m.Wrap(id)
	}

	return strings.Join(O, "\n")
}

func (m *Msg) Wrap(id string) string {
	return m.ID + ": " + m.BODY
}

type Bot struct {
	USER   *user.User
	PROMPT string
}
