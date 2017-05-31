package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"strings"
	"time"
	"unicode"

	"gopkg.in/telegram-bot-api.v4"
)

// Config holds configuration variables essentials to the execution of the program.
type Config struct {
	BotToken string `json:"botToken"`
}

// NewConfig reads a JSON file at fp, unmarshals it and returns a corresponding Config.
func NewConfig(fp string) (Config, error) {
	buf, err := ioutil.ReadFile(fp)
	c := Config{}

	if err != nil {
		return c, err
	}

	err = json.Unmarshal(buf, &c)
	if err != nil {
		return Config{}, err
	}

	return c, nil
}

func printError(reason string, err error) {
	if err != nil {
		log.Fatalf("%s: %v", reason, err)
	}
}

var conf Config
var bot *tgbotapi.BotAPI

func mangle(str string) (ret string) {
	ret = ""
	for _, word := range strings.Split(str, " ") {
		newword := ""
		rw := []rune(word)
		if len(rw) == 0 {
			continue
		}
		// the word without the last character
		ws := string(rw[0:(len(rw) - 1)])
		for i := 0; i < len(rw); i++ {
			rnd := rand.Int() % 2
			r := string(rw[i])
			if rnd == 1 ||
				(i == len(rw)-1 && (ws == newword)) {
				if rw[i] == unicode.ToUpper(rw[i]) {
					newword += string(unicode.ToLower(rw[i]))
				} else {
					newword += string(unicode.ToUpper(rw[i]))
				}
			} else {
				newword += r
			}
		}
		ret += newword + " "
	}

	return
}

func main() {
	conf, err := NewConfig("config.json")
	printError("Couldn't read config", err)

	bot, err = tgbotapi.NewBotAPI(conf.BotToken)
	printError("Couldn't establish telegram bot api connection", err)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	printError("Coudln't get telegram bot updates", err)

	results := make([]interface{}, 1)
	for update := range updates {
		if update.InlineQuery != nil {
			toMangle := update.InlineQuery.Query
			mangled := mangle(toMangle)
			if mangled == "" || err != nil {
				if err != nil {
					log.Println(err.Error())
				}
				continue
			}
			doc := tgbotapi.NewInlineQueryResultArticle(time.Now().Format("mangle_%s"), "Mangled text", mangled)
			doc.Description = mangled
			results[0] = doc
			ic := tgbotapi.InlineConfig{}
			ic.InlineQueryID = update.InlineQuery.ID
			ic.Results = results
			ic.IsPersonal = false
			ic.CacheTime = 0
			_, err = bot.AnswerInlineQuery(ic)
			if err != nil {
				log.Printf("Couldn't answer to inline query: %v\n", err)
			}
		}
	}
}
