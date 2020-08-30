package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

func MainHandler(m *tb.Message, b *tb.Bot) {
	log.Println("message: ", m)
	b.Send(m.Sender, "[WIP] 乱七八糟功能的bot")
}

func ArchiveHandler(m *tb.Message, b *tb.Bot) {
	log.Println("message: ", m.Text, "payload: ", m.Payload)
	b.Send(m.Sender, "未完成 ")
}

func UrbanDictHandler(m *tb.Message, b *tb.Bot) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			b.Send(m.Sender, fmt.Sprintln("Error: ", err))
		}
	}()

	log.Println(m.Text, m.Text)
	if m.Payload == "" {
		return
	}
	text, dictList := QueryKeywordFromUrAPI(m.Payload)

	_, err := b.Send(m.Sender, text, "HTML")
	if err != nil {
		log.Println(err)
		b.Send(m.Sender, fmt.Sprintln("Error: ", err.Error(), "message length: ", len(text)))
		if len(dictList) > 0 {
			b.Send(m.Sender, dictList[0].Permalink)
		} else {
			b.Send(m.Sender, "No result.")
		}

	}
}

func UrbanDictInLineQueryHander(q *tb.Query, b *tb.Bot) {
	log.Println(q.From, q.Text)
	_, dictList := QueryKeywordFromUrAPI(q.Text)
	if len(dictList) == 0 {
		return
	}
	results := make(tb.Results, len(dictList)) // []tb.Result
	for i, urRes := range dictList {
		result := &tb.ArticleResult{
			Title: urRes.Word + ": " + strconv.Itoa(i+1) + " " + urRes.Example,
			Text:  urRes.Definition + "\n\n" + urRes.Example + "\n\n" + urRes.Permalink,
		}

		results[i] = result
		// needed to set a unique string ID for each result
		results[i].SetResultID(strconv.Itoa(i))
	}
	err := b.Answer(q, &tb.QueryResponse{
		Results:   results,
		CacheTime: 60, // a minute
	})

	if err != nil {
		log.Println(err)
	}
}

func TwiiterSearchHandler(m *tb.Message, b *tb.Bot) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			b.Send(m.Sender, fmt.Sprintln("Error: ", err))
		}
	}()
	log.Println("message: ", m)
	tweets := SearchTweets(m.Payload).Statuses
	if len(tweets) == 0 {
		b.Send(m.Sender, "No tweets", "HTML")
	} else {
		b.Send(m.Sender, formatTweets(tweets), "HTML")
	}
}

func StockInfoHandler(m *tb.Message, b *tb.Bot) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			b.Send(m.Sender, fmt.Sprintln("Error: ", err))
		}
	}()
	log.Println("message: ", m)
	symbol := strings.ToUpper(m.Payload)
	url := JudgeStockCountryApiUrl(symbol)
	info := QueryStockInfoFromAPI(url)
	photo := DrawStockPricesCharts(info)
	b.Send(m.Sender, photo)

}
