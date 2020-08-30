// tgbotgo project main.go
package main

import (
	"log"
	"time"

	"github.com/spf13/viper"
	tb "gopkg.in/tucnak/telebot.v2"
)

func initViper() {
	viper.SetConfigName("conf")
	viper.AddConfigPath(".")
	viper.SetConfigType("toml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}
	log.Println("viper.AllKeys", viper.AllKeys())
}

func main() {
	initViper()
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	b, err := tb.NewBot(tb.Settings{
		Token:  viper.GetString("telegram.bot_token"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	log.Println("bot", b)
	if err != nil {
		log.Println(err)
		return
	}

	b.Handle("/save", func(m *tb.Message) {
		go ArchiveHandler(m, b)
	})

	b.Handle("/ur", func(m *tb.Message) {
		go UrbanDictHandler(m, b)
	})

	b.Handle("/tw", func(m *tb.Message) {
		go TwiiterSearchHandler(m, b)
	})

	b.Handle("/s", func(m *tb.Message) {
		go StockInfoHandler(m, b)
	})

	b.Handle(tb.OnQuery, func(q *tb.Query) {
		go UrbanDictInLineQueryHander(q, b)
	})

	b.Handle(tb.OnText, func(m *tb.Message) {
		go MainHandler(m, b)
	})

	b.Start()
}
