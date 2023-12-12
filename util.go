package main

import (
	"strings"
	"unicode"

	"github.com/rs/zerolog/log"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Removes Unicode accents from string.
func removeAccents(s string) string {

	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	o, _, e := transform.String(t, s)
	if e != nil {
		log.Fatal().Err(e).Send()
	}

	return o
}

// Replicate API POST request.
type ReqR struct {
	Version string      `json:"version"`
	Input   interface{} `json:"input"`
}

type ReqRLLaMa struct {
	Prompt            string  `json:"prompt"`
	SystemPrompt      string  `json:"system_prompt"`
	MaxNewTokens      int     `json:"max_new_tokens"`
	MinNewTokens      int     `json:"min_new_tokens"`
	Temperature       float64 `json:"temperature"`
	RepetitionPenalty float64 `json:"repetition_penalty"`
	TopK              int     `json:"top_k"`
	TopP              float64 `json:"top_p"`
}

// Replicate API POST response.
type ResR struct {
	Output []string `json:"output"`
}

type Bot struct {
	USER   *User
	PROMPT string
}

func (B *Bot) String() string {
	return strings.Join([]string{
		"ID: " + B.USER.ID,
		"COLOR: " + B.USER.COLOR,
		"PROMPT: " + B.PROMPT,
		"DONE",
	}, "\n")
}
