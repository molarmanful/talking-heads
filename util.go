package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"unicode"

	nanoid "github.com/matoous/go-nanoid/v2"
	"github.com/olahol/melody"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func calcWs(lms []*Msg) []*Bot {

	ws := make(map[string]int)
	for _, m := range lms {
		if v, ok := wbots[m.ID]; ok {
			ws = v
		}
	}
	if ws == nil {
		log.Warn().Msg("no usr msgs?")
		return bots
	}

	for i, m := range lms {
		cn := CM.ClosestN(strings.ToLower(removeAccents(m.BODY)), 5)
		for j, id := range cn {
			ws[id] += max(0, len(bots)/2+10-i*i-j*j)
		}

		if w, ok := ws[m.ID]; ok {
			w += len(bots) / 2
		}
	}

	bs := make([]*Bot, 0, len(bots)*2)
	for id, n := range ws {
		for j := 0; j < n; j++ {
			bs = append(bs, botmap[id])
		}
	}

	return bs
}

func postReq(M *melody.Melody, bot *Bot, relstr string) (string, error) {

	id := bot.USER.ID
	log.Info().Msg("Q: " + id)

	r := tagMsgs(id, 7)
	j, e := json.Marshal(map[string]interface{}{
		// "version": "02e509c789964a7ea8736978a43525956ef40397be9033abf9fd2badfe68c9e3", // llama 2 70b
		"version": "f4e2de70d66816a838a89eeeb621910adffb0dd0baba3976c96980970978018d", // llama 2 13b
		"input": map[string]interface{}{
			"top_k":       50,
			"top_p":       .95,
			"prompt":      r,
			"temperature": .8,
			"system_prompt": strings.Join([]string{
				bot.PROMPT,
				relstr, "\n",
				`Generate a concise one-sentence response to the conversation as ` + id + `, without using speaker labels and ensuring relevance to the context provided.`,
				`Example responses:\nWitness my power, mere mortal!\nYou will suffer for your transgressions, NPC#F69420.\nZEUS, I find you tolerable.`,
			}, " "),
			"max_new_tokens":     100,
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

	M.Broadcast([]byte(bot.USER.MkMsg("+t", "")))
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

func tagMsgs(id string, n int) string {

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

func relStr(lms []*Msg, id string) string {

	uniq := make(map[string]bool)
	rs := make([]string, 0, 10)
	for _, m := range lms {
		if v, ok := rels[m.ID]; ok {
			if _, ok := uniq[m.ID]; ok {
				continue
			}
			uniq[m.ID] = true

			s := "tolerate"
			if v[id] >= 75 {
				s = "greatly treasure"
			} else if v[id] >= 50 {
				s = "are fond of"
			} else if v[id] >= 25 {
				s = "like"
			} else if v[id] >= 0 {
				s = "are neutral towards"
			} else if v[id] <= -75 {
				s = "absolutely hate"
			} else if v[id] <= -50 {
				s = "are disgusted by"
			} else if v[id] <= -25 {
				s = "dislike"
			}
			rs = append(rs, s+" "+m.ID)
		}
	}

	return "You " + strings.Join(rs, ", ") + "."
}

func removeAccents(s string) string {

	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	o, _, e := transform.String(t, s)
	if e != nil {
		panic(e)
	}

	return o
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
