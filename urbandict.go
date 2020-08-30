package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/viper"
)

type UrbanDictItem struct {
	Definition string `json:"definition"`
	Permalink  string `json:"permalink"`
	Author     string `json:"author"`
	Word       string `json:"word"`
	WriteOn    string `json:"written_on"`
	Example    string `json:"example"`
}

type UrbanJsonResult struct {
	List []UrbanDictItem `json:"list"`
}

func QueryKeywordFromUrAPI(keyword string) (string, []UrbanDictItem) {
	resp, err := http.Get(viper.GetString("urbandict.dict_api_url") + keyword)
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("status code error: %d %s %s", resp.StatusCode, resp.Status, resp.Header)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("read body err, %v\n", err)
		return err.Error(), []UrbanDictItem{}
	}
	log.Printf("%s", body)

	var urResult UrbanJsonResult
	if err = json.Unmarshal(body, &urResult); err != nil {
		log.Printf("Unmarshal err, %v\n", err)
		return err.Error(), []UrbanDictItem{}
	}
	textList := urResult.List
	text := formatDictResult(urResult)
	return text, textList
}

func formatDictResult(urResult UrbanJsonResult) string {
	text := ""
	for _, item := range urResult.List {
		textPiece := fmt.Sprintf(
			"<a href=\"%s\"><b>%s</b></a> \n \n"+
				"<b>Definition</b>: \n  %s \n"+
				" \n"+
				"<i><b>Example</b>: \n %s</i> \n"+
				"\n \n \n",
			item.Permalink, item.Word, item.Definition, item.Example)

		text += textPiece
	}
	text = strings.ReplaceAll(text, "[", "<b>")
	text = strings.ReplaceAll(text, "]", "</b>")
	return text
}
