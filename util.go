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

// Generates bot weights based on recent msgs.
func calcWs(lms []*Msg) ([]*Bot, string) {

	// get last user + their weights
	ws := make(map[string]int)
	lU := ""
	for _, m := range lms {
		if v, ok := wbots[m.USER.ID]; ok {
			ws = v
			lU = m.USER.ID
		}
	}
	if ws == nil {
		log.Warn().Msg("no usr msgs?")
		return bots, lU
	}

	for i, m := range lms {
		// get mentioned bots
		// may gen false positives, but that is fine
		cn := CM.ClosestN(strings.ToLower(removeAccents(m.BODY)), 5)
		for j, id := range cn {
			ws[id] += max(0, len(bots)/2+10-i*i-j*j)
		}

		// add weight to bot w/ recent msg
		if w, ok := ws[m.USER.ID]; ok {
			w += len(bots) / 2
		}
	}

	// replicate bots into weighted list
	bs := make([]*Bot, 0, len(bots)*2)
	for id, n := range ws {
		for j := 0; j < n; j++ {
			bs = append(bs, botmap[id])
		}
	}

	return bs, lU
}

// Synchronously requests response from Replicate API proxy.
func reqRes(M *melody.Melody, bot *Bot, relstr string) (string, error) {

	id := bot.USER.ID
	log.Info().Msg("Q: " + id)

	r := tagMsgs(id, conf.PLastN)
	j, e := json.Marshal(&ReqR{
		// Version: "02e509c789964a7ea8736978a43525956ef40397be9033abf9fd2badfe68c9e3", // llama 2 70b
		Version: "f4e2de70d66816a838a89eeeb621910adffb0dd0baba3976c96980970978018d", // llama 2 13b
		Input: &ReqRInput{
			Prompt: r,
			SystemPrompt: strings.Join([]string{
				bot.PROMPT,
				relstr, "\n",
				`Generate a concise one-sentence response to the conversation as ` + id + `, without using speaker labels and ensuring relevance to the context provided.`,
				`If you understand the prompt, start your response with "RES:".`,
				`Example responses:\nRES: Witness my power, mere mortal!\nRES: You will suffer for your transgressions, NPC#F69420.\nRES: ZEUS, I find you tolerable.`,
			}, " "),
			MaxNewTokens:      100,
			MinNewTokens:      -1,
			Temperature:       .9,
			RepetitionPenalty: 1.18,
			TopK:              30,
			TopP:              .73,
		},
	})
	if e != nil {
		return "", e
	}

	// req
	req, e := http.NewRequest("POST", "https://replicate-api-proxy.glitch.me/create_n_get", bytes.NewBuffer(j))
	if e != nil {
		return "", e
	}
	req.Header.Add("Content-Type", "application/json")

	// res
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

	post := &ResR{}
	if e = json.NewDecoder(res.Body).Decode(post); e != nil {
		return "", e
	}

	O := strings.TrimSpace(strings.Join(post.Output, ""))
	log.Info().Msg(O)
	return strings.TrimSpace(strings.TrimPrefix(O, "RES:")), nil
}

// Converts last n messages to tagged prompt.
// Non [bot w/ id] answers are tagged.
func tagMsgs(id string, n int) string {

	n = min(len(msgs), n)
	O := ""
	ins := false
	for _, m := range msgs[len(msgs)-n:] {
		if m.USER.ID == id {
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
		O += m.USER.ID + ": " + m.BODY + "\n"
		ins = true
	}
	if ins {
		O = O[:len(O)-1] + " [/INS]\n"
	}

	return O
}

// Converts relations w/ recent users to prompt.
func relStr(lms []*Msg, id string) string {

	uniq := make(map[string]bool)
	rs := make([]string, 0, 10)
	for _, m := range lms {
		if v, ok := rels[m.USER.ID]; ok {
			if _, ok := uniq[m.USER.ID]; ok {
				continue
			}
			uniq[m.USER.ID] = true

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
			rs = append(rs, s+" "+m.USER.ID)
		}
	}

	return "You " + strings.Join(rs, ", ") + "."
}

// Remove Unicode accents from string.
func removeAccents(s string) string {

	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	o, _, e := transform.String(t, s)
	if e != nil {
		panic(e)
	}

	return o
}

// Replicate API POST request.
type ReqR struct {
	Version string     `json:"version"`
	Input   *ReqRInput `json:"input"`
}

type ReqRInput struct {
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

type Msg struct {
	USER *User
	BODY string
}

// Get msg as string.
func (m *Msg) String() string {
	return m.USER.String() + " " + m.BODY
}

type Bot struct {
	USER   *User
	PROMPT string
}

type User struct {
	ID    string
	COLOR string
}

// Gen new user.
func NewUser() *User {

	id, e := nanoid.Generate("0123456789ABCDEF", 6)
	if e != nil {
		panic(e)
	}
	return &User{"NPC#" + id, "#" + id}
}

// Get user as string.
func (u *User) String() string {
	return u.ID + " " + u.COLOR
}

// Gen msg from header + user + body.
func (u *User) MkMsg(h string, b string) string {
	return h + " " + u.String() + " " + b
}
