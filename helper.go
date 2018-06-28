package main

import (
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

var symbols = []string{
	"ETHBTC", "LTCBTC", "XMRBTC", "XRPBTC", "BNBBTC", "NEOBTC",
	"1ETHBTC", "1LTCBTC", "1XMRBTC", "1XRPBTC", "1BNBBTC", "1NEOBTC",
	"2ETHBTC", "2LTCBTC", "2XMRBTC", "2XRPBTC", "2BNBBTC", "2NEOBTC",
	"3ETHBTC", "3LTCBTC", "3XMRBTC", "3XRPBTC", "3BNBBTC", "3NEOBTC",
	"4ETHBTC", "4LTCBTC", "4XMRBTC", "4XRPBTC", "4BNBBTC", "4NEOBTC",
	"5ETHBTC", "5LTCBTC", "5XMRBTC", "5XRPBTC", "5BNBBTC", "5NEOBTC",
	"6ETHBTC", "6LTCBTC", "6XMRBTC", "6XRPBTC", "6BNBBTC", "6NEOBTC",
	"7ETHBTC", "7LTCBTC", "7XMRBTC", "7XRPBTC", "7BNBBTC", "7NEOBTC",
	"8ETHBTC", "8LTCBTC", "8XMRBTC", "8XRPBTC", "8BNBBTC", "8NEOBTC",
	"9ETHBTC", "9LTCBTC", "9XMRBTC", "9XRPBTC", "9BNBBTC", "9NEOBTC",
	"0ETHBTC", "0LTCBTC", "0XMRBTC", "0XRPBTC", "0BNBBTC", "0NEOBTC",
	"1XTHBTC", "1XTCBTC", "1ZMRBTC", "1ZRPBTC", "1XNBBTC", "1XEOBTC",
	"2XTHBTC", "2XTCBTC", "2ZMRBTC", "2ZRPBTC", "2XNBBTC", "2XEOBTC",
	"3XTHBTC", "3XTCBTC", "3ZMRBTC", "3ZRPBTC", "3XNBBTC", "3XEOBTC",
	"4XTHBTC", "4XTCBTC", "4ZMRBTC", "4ZRPBTC", "4XNBBTC", "4XEOBTC",
	"5XTHBTC", "5XTCBTC", "5ZMRBTC", "5ZRPBTC", "5XNBBTC", "5XEOBTC",
	"6XTHBTC", "6XTCBTC", "6ZMRBTC", "6ZRPBTC", "6XNBBTC", "6XEOBTC",
}

type Ticker struct {
	ID        int
	Symbol    string
	Price     float64
	Timestamp time.Time
}

func TickerStream() chan []*Ticker {
	ch := make(chan []*Ticker)
	go func(ch chan []*Ticker) {
		buffer := make([][]*Ticker, 0, genratorStreamCount)
		for i := 0; i < genratorStreamCount; i++ {
			//ch <- generateRandomTickers(genratorArrSize)
			//log.Println("Done: ", i)
			buffer = append(buffer, generateRandomTickers(genratorArrSize))
		}
		log.Println("Begining sending")
		for i := 0; i < genratorStreamCount; i++ {
			ch <- buffer[i]
		}
		close(ch)
	}(ch)
	return ch
}

func generateRandomTickers(count int) []*Ticker {
	dSyms := make([]string, count, count)
	copy(dSyms, symbols)
	tArr := make([]*Ticker, 0, count)
	for i := 0; i < count; i++ {
		s := ""
		dSymsLen := len(dSyms)
		if dSymsLen > 0 {
			idx := rand.Intn(dSymsLen)
			s = dSyms[idx]
			dSyms = append(dSyms[:idx], dSyms[idx+1:]...)
		} else {
			log.Println("WHAT!", len(tArr))
			os.Exit(1)
		}
		tArr = append(tArr, NewRandomTicker(s))
	}
	return tArr
}

// marketInfo stores data needed for pump detection
type marketInfo struct {
	buffer              []*Ticker
	lastNonZeroVelocity float64
	pdCxt               *pdContext
}

// newMarketInfo new instance
func newMarketInfo() *marketInfo {
	return &marketInfo{make([]*Ticker, 0, 3), 0, nil}
}

func NewRandomTicker(symbol string) *Ticker {
	id++
	s := symbol
	if len(s) == 0 {
		cnt := len(symbols)
		s = symbols[rand.Intn(cnt)]
	}
	p := 0.3432
	t := time.Unix(int64(id), 0)
	return &Ticker{id, s, p, t}
}

var id int = -1

// pdContext encompass values needed for executions of p&d trade
type pdContext struct {
	symbol       string
	triggerPrice float64
	buyPrice     float64
	sellPrice    float64
	base         string
	quote        string
}

// newPDContext new instance of pdContext
func newPdContext(t *Ticker) *pdContext {
	quote := "BTC"
	base := strings.Replace(t.Symbol, quote, "", -1)
	return &pdContext{
		symbol:       t.Symbol,
		triggerPrice: t.Price,
		buyPrice:     t.Price * 1.04,
		sellPrice:    t.Price * 1.11,
		base:         base,
		quote:        quote,
	}
}

func hasUniqueElements(arr []*Ticker) bool {
	seenMarkets := make(map[string]bool)
	for i, t := range arr {
		if _, ok := seenMarkets[t.Symbol]; ok {
			log.Println("First dumplicate index", i)
			logEnumerate(arr)
			return false
		}
		seenMarkets[t.Symbol] = true
	}
	return true
}

func logEnumerate(tArr []*Ticker) {
	for i, t := range tArr {
		log.Println(i, *t)
	}
}
