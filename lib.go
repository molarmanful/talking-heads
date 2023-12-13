package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	redis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// Connects to Redis
func (ST *State) InitR(url string) *State {

	opt, e := redis.ParseURL(url)
	if e != nil {
		log.Fatal().Err(e).Msg("redis connect failed " + url)
	}
	ST.R = redis.NewClient(opt)
	return ST
}

// Generates bot weights based on recent msgs.
func (ST *State) CalcWs(lms []*Msg) ([]*Bot, string) {

	// get last user + their weights
	ws := make(map[string]int)
	lU := ""
	for _, m := range lms {
		if v, ok := ST.Wbots[m.USER.ID]; ok {
			ws = v
			lU = m.USER.ID
		}
	}
	if ws == nil {
		log.Warn().Msg("no usr msgs?")
		return ST.Bots, lU
	}

	for i, m := range lms {
		// get mentioned bots
		// may gen false positives, but that is fine
		cn := ST.CM.ClosestN(strings.ToLower(removeAccents(m.BODY)), 5)
		for j, id := range cn {
			ws[id] += max(0, len(ST.Bots)/2+10-i*i-j*j)
		}

		// add weight to bot w/ recent msg
		if w, ok := ws[m.USER.ID]; ok {
			w += len(ST.Bots) / 2
		}
	}

	// replicate bots into weighted list
	bs := make([]*Bot, 0, len(ST.Bots)*2)
	for id, n := range ws {
		for j := 0; j < n; j++ {
			bs = append(bs, ST.BotMap[id])
		}
	}

	return bs, lU
}

// Synchronously requests response from Replicate API proxy.
func (ST *State) ReqRes(j []byte, during func(*http.Request)) (string, error) {

	// req
	req, e := http.NewRequest("POST", "https://replicate-api-proxy.glitch.me/create_n_get", bytes.NewBuffer(j))
	if e != nil {
		return "", e
	}
	req.Header.Add("Content-Type", "application/json")

	// res
	during(req)
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
	return O, nil
}

// Synchronously requests god response from Replicate API proxy.
func (ST *State) ReqResGod(bot *Bot, relstr string) (string, error) {

	id := bot.USER.ID
	log.Info().Msg("Q: " + id)
	r := ST.TagMsgs(id, rand.Intn(ST.PLastN)+1)

	j, e := json.Marshal(&ReqR{
		// Version: "02e509c789964a7ea8736978a43525956ef40397be9033abf9fd2badfe68c9e3", // llama 2 70b chat
		Version: "f4e2de70d66816a838a89eeeb621910adffb0dd0baba3976c96980970978018d", // llama 2 13b chat
		Input: &ReqRLLaMa{
			Prompt: r,
			SystemPrompt: strings.Join([]string{
				bot.PROMPT,
				relstr, "\n",
				`Generate a concise one-sentence response as ` + id + ` to any message in the conversation,`,
				`without using speaker labels and ensuring relevance to the context provided.`,
				`If you understand this prompt, start your response with "RES:".`,
				`Example responses:\nRES: Witness my power, mere mortal!\nRES: You will suffer for your transgressions, NPC#F69420.\nRES: ZEUS, I find you tolerable.`,
			}, " "),
			MaxNewTokens:      128,
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

	O, e := ST.ReqRes(j, func(req *http.Request) {
		ST.M.Broadcast([]byte(bot.USER.MkMsg("+t", "")))
	})
	if e != nil {
		return "", e
	}

	log.Info().Msg(O)
	O = strings.TrimSpace(strings.TrimPrefix(O, "RES:"))
	return O, nil
}

// Synchronously requests sentiment analysis response from Replicate API proxy.
func (ST *State) ReqResFeels(s string) (float64, error) {

	log.Info().Msg("FEEL: " + s)

	j, e := json.Marshal(&ReqR{
		Version: "13c3cdee13ee059ab779f0291d29054dab00a47dad8261375654de5540165fb0", // llama 2 7b chat
		Input: &ReqRLLaMa{
			Prompt: s,
			SystemPrompt: strings.Join([]string{
				"You are an accurate sentiment analyzer.",
				"Given a message sent from a god to a mortal, your job is to analyze how the god feels about the mortal with:",
				"-3 for hate, -2 for dislike, -1 for mild dislike, 0 for neutral, 1 for mild like, 2 for like, 3 for love.",
				"The first line of your response is the number alone.",
				"The second line of your response is a concise reason for your analysis.",
			}, " "),
			MaxNewTokens:      128,
			MinNewTokens:      -1,
			Temperature:       .3,
			RepetitionPenalty: 1,
			TopK:              -1,
			TopP:              .95,
		},
	})
	if e != nil {
		return 0, e
	}

	O, e := ST.ReqRes(j, func(req *http.Request) {})
	if e != nil {
		return 0, e
	}
	log.Info().Msg(O)

	numR := regexp.MustCompile(`[\d.-]+`)

	n, e := strconv.ParseFloat(numR.FindString(O), 64)
	if e != nil {
		return 0, e
	}

	return max(-1, min(1, n/3)), nil
}

// Synchronously requests new god profile from Replicate API proxy.
func (ST *State) ReqResNewGod() (*Bot, error) {

	j, e := json.Marshal(&ReqR{
		Version: "f4e2de70d66816a838a89eeeb621910adffb0dd0baba3976c96980970978018d", // llama 2 13b chat
		Input: &ReqRLLaMa{
			Prompt: "",
			SystemPrompt: strings.Join([]string{
				"You are a god creation machine.",
				"The following gods have been created:",
				strings.Join(ST.BotNames(), ", ") + ".",
				"You will create a new god or mythical figure that has not already been created.",
        "Draw from a wide variety of cultures, domains, and personalities.",
        "Examples:\n---\n\n",
				ST.BotExamples() + "\n\n---\n",
				"You are obedient; you will strictly follow the above example structures and say nothing else.",
				"Say DONE when you are done and have complied with the specifications.",
			}, " "),
			MaxNewTokens:      256,
			MinNewTokens:      -1,
			Temperature:       .9,
			RepetitionPenalty: 1.18,
			TopK:              30,
			TopP:              .73,
		},
	})
	if e != nil {
		return nil, e
	}

	O, e := ST.ReqRes(j, func(req *http.Request) {})
	if e != nil {
		return nil, e
	}
	log.Info().Msg(O)

	b, e := ST.AddBot(O)
	if e != nil {
		return nil, e
	}

	return b, nil
}

// Gets list of bot names.
func (ST *State) BotNames() []string {

	bn := make([]string, 0, len(ST.BotMap))
	for k := range ST.BotMap {
		bn = append(bn, k)
	}

	return bn
}

// Gets bot example prompts.
func (ST *State) BotExamples() string {

	bs := ST.Bots[0:5]
	es := make([]string, 5)
	for i, b := range bs {
		es[i] = b.String()
	}

	return strings.Join(es, "\n\n")
}

// Parses prompt and adds bot.
func (ST *State) AddBot(s string) (*Bot, error) {

	linesR := regexp.MustCompile("\r?\n+")
	bot := &Bot{&User{"", ""}, ""}
	for _, l := range linesR.Split(s, -1) {
		if strings.HasPrefix(l, "ID: ") {
			bot.USER.ID = strings.TrimPrefix(l, "ID: ")
		} else if strings.HasPrefix(l, "COLOR: ") {
			bot.USER.COLOR = strings.TrimPrefix(l, "COLOR: ")
		} else if strings.HasPrefix(l, "PROMPT: ") {
			bot.PROMPT = strings.TrimPrefix(l, "PROMPT: ")
		}
	}
	if bot.USER.ID == "" || bot.USER.COLOR == "" || bot.PROMPT == "" {
		return nil, errors.New("bad bot profile")
	}
	if _, ok := ST.BotMap[bot.USER.ID]; ok {
		return nil, errors.New("bot already exists")
	}

	ST.BotsMu.Lock()
	ST.BotMapMu.Lock()
	ST.WbotsMu.Lock()
	ST.RelsMu.Lock()

	ST.Bots = append(ST.Bots, bot)
	ST.BotsMu.Unlock()

	ST.BotMap[bot.USER.ID] = bot
	ST.BotMapMu.Unlock()

	for u := range ST.Wbots {
		ST.Wbots[u][bot.USER.ID] = 1
		ST.Rels[u][bot.USER.ID] = float64(rand.Intn(100) - 50)
	}
	ST.WbotsMu.Unlock()
	ST.RelsMu.Unlock()

	go ST.StoreGod(bot)

	return bot, nil
}

// Converts last n messages to tagged prompt.
// Non [bot w/ id] answers are tagged.
func (ST *State) TagMsgs(id string, n int) string {

	n = min(len(ST.Msgs), n)
	O := ""
	ins := false
	for _, m := range ST.Msgs[len(ST.Msgs)-n:] {
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
func (ST *State) RelStr(lms []*Msg, id string) string {

	uniq := make(map[string]bool)
	rs := make([]string, 0, 10)
	for _, m := range lms {
		if v, ok := ST.Rels[m.USER.ID]; ok {
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

// Gets msgs from Redis.
func (ST *State) GetMsgs() ([]*Msg, error) {

	res, e := ST.R.LRange(ST.Ctx, "th:chat", 0, -1).Result()
	if e != nil {
		return nil, e
	}

	ms := make([]*Msg, len(res))
	for i, v := range res {
		ms[i] = ToMsg(v)
	}

	return ms, nil
}

// Gets gods from Redis.
func (ST *State) GetGods() error {

	res, e := ST.R.LRange(ST.Ctx, "th:gods", 0, -1).Result()
	if e != nil {
		return e
	}

	for _, v := range res {
		ST.AddBot(v)
	}

	return nil
}

// Store msgs in Redis.
func (ST *State) StoreMsg(m *Msg) {
	if e := ST.R.RPush(ST.Ctx, "th:chat", m.String()).Err(); e != nil {
		log.Error().Err(e).Msg("redis write th:chat error")
	}
}

// Store gods in Redis.
func (ST *State) StoreGod(b *Bot) {
	if e := ST.R.RPush(ST.Ctx, "th:gods", b.String()).Err(); e != nil {
		log.Error().Err(e).Msg("redis write th:gods error")
	}
}
