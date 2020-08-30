package main

import (
	"fmt"
	"log"

	"github.com/spf13/viper"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

func CreateTwitterClient() *twitter.Client {
	initViper()
	ConsumerKey := viper.GetString("twitter.consumer_key")
	ConsumerSecret := viper.GetString("twitter.consumer_secret")
	AccessToken := viper.GetString("twitter.access_token")
	AccessSecret := viper.GetString("twitter.access_secret")

	config := oauth1.NewConfig(ConsumerKey, ConsumerSecret)
	log.Println("config: ", config)
	token := oauth1.NewToken(AccessToken, AccessSecret)
	log.Println("token: ", token)
	// OAuth1 http.Client will automatically authorize Requests
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	client := twitter.NewClient(httpClient)
	return client
}

var TwitterClient *twitter.Client = CreateTwitterClient()

func SearchTweets(keyword string) *twitter.Search {
	search, _, err := TwitterClient.Search.Tweets(&twitter.SearchTweetParams{
		Query:      keyword,
		Count:      10,
		ResultType: "mixed",
		Since:      "2012-01-01",
	})
	if err != nil {
		log.Println(err)
	}
	return search
}

func formatTweets(tweets []twitter.Tweet) string {
	log.Println("Size", len(tweets))
	text := ""
	for _, tweet := range tweets {
		textPiece := fmt.Sprintf(
			"<a href=\"https://twitter.com/%s/status/%s\"> <b>%s</b>  </a>: \n"+
				" %s \n\n\n", tweet.User.ScreenName, tweet.IDStr, tweet.User.Name, tweet.Text)

		text += textPiece
	}

	return text
}
