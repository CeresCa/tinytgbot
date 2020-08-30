package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/viper"
	tb "gopkg.in/tucnak/telebot.v2"

	chart "github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

type HourTrading struct {
	Tag         string  `json:"tag"`
	LatestPrice float64 `json:"latestPrice"`
	PreClose    float64 `json:"preClose"`
	LatestTime  string  `json:"latestTime"`
	Volume      int64   `json:"volume"`
	Timestamp   int64   `json:"timestamp"`
}

type StockDetail struct {
	LatestPrice   float64     `json:"latestPrice"`
	Halted        float32     `json:"halted"`
	Amount        float64     `json:"amount"`
	Change        float64     `json:"change"`
	NameCN        string      `json:"nameCN"`
	TradingStatus int8        `json:"tradingStatus"`
	Volume        int64       `json:"volume"`
	High          float64     `json:"high"`
	Amplitude     float64     `json:"amplitude"`
	AdjPreClose   float64     `json:"adjPreClose"`
	PreClose      float64     `json:"preClose"`
	Low           float64     `json:"low"`
	MarketStatus  string      `json:"marketStatus"`
	Exchange      string      `json:"exchange"`
	LatestTime    string      `json:"latestTime"`
	Open          float64     `json:"open"`
	Timestamp     int64       `json:"timestamp"`
	HourTrading   HourTrading `json:"hourTrading"`
}

type StockPrice struct {
	Volume    float64 `json:"volume"`
	Price     float64 `json:"price"`
	AvgPrice  float64 `json:"avgPrice"`
	Timestamp int64   `json:"time"`
}

type StockInfo struct {
	Detail   StockDetail  `json:"detail"`
	Items    []StockPrice `json:"items"`
	Symbol   string       `json:"symbol"`
	PreClose float64      `json:"preClose"`
}

func QueryStockInfoFromAPI(url string) StockInfo {
	log.Println("url: ", url)
	resp, err := http.Get(url)
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
	}

	var LatestStockInfo StockInfo
	if err = json.Unmarshal(body, &LatestStockInfo); err != nil {
		log.Printf("Unmarshal err, %v\n", err)
	}
	log.Println(LatestStockInfo.Detail)
	return LatestStockInfo

}

func FormatStockInfo(info StockInfo) string {
	changeRate := (info.Detail.LatestPrice - info.PreClose) / info.PreClose * 100
	changeRateHour := (info.Detail.HourTrading.LatestPrice - info.Detail.HourTrading.PreClose) / info.Detail.HourTrading.PreClose * 100

	text := fmt.Sprintf("%s (%s)\t        名称: %s \n"+
		"现价: %.3f \n"+
		"涨跌额: %.3f\t        涨跌幅: %.2f%% \n"+
		"市场交易状态: %s \n\n", info.Symbol, info.Detail.Exchange, info.Detail.NameCN, info.Detail.LatestPrice, info.Detail.Change, changeRate, info.Detail.MarketStatus)

	if info.Detail.TradingStatus != 2 && info.Detail.Exchange != "SEHK" {
		text += fmt.Sprintf("盘前盘后\n现价：%.3f        涨跌幅：%.2f%%\n", info.Detail.HourTrading.LatestPrice, changeRateHour)
	}
	log.Println(text)
	return text
}

func DrawStockPricesCharts(info StockInfo) *tb.Photo {

	var prices []float64
	var volumes []float64
	var times []time.Time
	for _, v := range info.Items {
		prices = append(prices, v.Price)
		times = append(times, ConvertTimestamp(v.Timestamp))
		volumes = append(volumes, v.Volume)
	}

	priceSeries := chart.TimeSeries{
		Name: info.Symbol,
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.GetDefaultColor(0).WithAlpha(64),
			FillColor:   chart.GetDefaultColor(0).WithAlpha(64),
		},
		XValues: times,
		YValues: prices,
	}

	volumeSeries := chart.TimeSeries{
		Name: "Volume",
		Style: chart.Style{
			Show:        true,
			StrokeColor: drawing.ColorRed,
		},
		YAxis:   chart.YAxisSecondary,
		XValues: times,
		YValues: volumes,
	}

	graph := chart.Chart{
		Title: info.Symbol,
		TitleStyle: chart.Style{
			Show: true},
		Background: chart.Style{
			Padding: chart.Box{
				Top:    60,
				Left:   35,
				Right:  35,
				Bottom: 5,
			},
			FillColor: drawing.ColorFromHex("efefef"),
		},
		XAxis: chart.XAxis{
			Name: "Time",
			Style: chart.Style{
				Show: true},
			ValueFormatter: chart.TimeMinuteValueFormatter,
		},
		YAxis: chart.YAxis{
			Name: "Price",
			NameStyle: chart.Style{
				Show: true},
			Style: chart.Style{
				Show: true},
		},
		YAxisSecondary: chart.YAxis{
			Name: "Volume",
			NameStyle: chart.Style{
				Show: true},
			Style: chart.Style{
				Show: true},
			ValueFormatter: func(v interface{}) string {
				return fmt.Sprintf("%.f", v)
			},
		},
		Series: []chart.Series{
			priceSeries,
			volumeSeries,
		},
	}

	f := bytes.NewBuffer(nil)

	err := graph.Render(chart.PNG, f)
	if err != nil {
		log.Println(err)
	}
	photo := &tb.Photo{File: tb.FromReader(f), Caption: FormatStockInfo(info)}
	return photo
}

func ConvertTimestamp(timestampMillisecond int64) time.Time {
	tm := time.Unix(timestampMillisecond/1000, 0)
	return tm

}

func JudgeStockCountryApiUrl(symbol string) string {
	symbol = strings.ToUpper(symbol)
	if matched, _ := regexp.MatchString(`^\d+(?:\.SH)?$`, symbol); matched {
		if len(symbol) == 5 {
			return viper.GetString("stock.hk_stock_api_url") + symbol + "?manualRefresh=true"
		}
		return viper.GetString("stock.cn_stock_api_url") + symbol
	} else {
		if matched, _ := regexp.MatchString(`^HSI|HSCEI|HSCCI$`, symbol); matched {
			return viper.GetString("stock.hk_stock_api_url") + symbol + "?manualRefresh=true"
		}
		return viper.GetString("stock.us_stock_api_url") + symbol
	}

}
