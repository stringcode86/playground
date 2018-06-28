package main

import (
	"fmt"
	"log"
	"playground/goutils"
	"sync"
	"time"
)

const genratorArrSize int = 102
const genratorStreamCount int = 1000000

const bufferSize int = 3
const workerPoolSize int = 4

func main() {
	log.Println("Let the fun beging")
	sCh := TickerStream()
	pdt := NewPDTrader()
	pdt.SetStream(sCh)
	var input string
	fmt.Scanln(&input)
}

type PDTrader struct {
	markets       map[string]*marketInfo
	marketsMutext *sync.Mutex
	tArrCh        <-chan []*Ticker
}

func NewPDTrader() *PDTrader {
	return &PDTrader{
		make(map[string]*marketInfo, 0),
		&sync.Mutex{},
		nil}
}

func (pdt *PDTrader) SetStream(tArrCh <-chan []*Ticker) {
	pdt.tArrCh = tArrCh
	go pdt.processTickers()
}

func (pdt *PDTrader) processTickers() {
	for {
		select {
		case tArr, ok := <-pdt.tArrCh:
			// Quit if channel is closed
			if ok == false {
				log.Println("Channel closed")
				return
			}
			pdt.dispatchPumpSerachWork3(tArr)
			// TODO: Remove return
			//log.Println("We are done")
			//return
		}
	}
}

func (pdt *PDTrader) dispatchPumpSerachWork(tArr []*Ticker) {
	wCh := make(chan *workerItem)
	rCh := make(chan string, workerPoolSize+1)
	wIDs := make(map[string]bool)
	jobTotal := len(tArr)
	// Create workers
	for i := 0; i < workerPoolSize; i++ {
		go pdt.worker(i, wCh, rCh)
	}
	for {
		select {
		case rID := <-rCh:
			delete(wIDs, rID)
			jobTotal--
			if jobTotal == 0 {
				close(wCh)
				return
			}
		default:
			if len(tArr) == 0 {
				continue
			}
			t := tArr[0]
			if _, ok := wIDs[t.Symbol]; ok {
				continue
			}
			tArr = tArr[1:]
			wIDs[t.Symbol] = true
			mi := pdt.marketInfo(t.Symbol)
			wi := &workerItem{t, mi}
			wCh <- wi // &workerItem{t, mi}
		}
	}
}

func (pdt *PDTrader) dispatchPumpSerachWork2(tArr []*Ticker) {
	for _, t := range tArr {
		mi := pdt.marketInfo(t.Symbol)
		pdt.searchPump(t, mi)
	}
}

func (pdt *PDTrader) worker(id int, wCh chan *workerItem, rCh chan string) {
	for {
		wi, ok := <-wCh
		if ok == false {
			return
		}
		pdt.searchPump(wi.t, wi.mi)
		rCh <- wi.t.Symbol
	}
}

func (pdt *PDTrader) marketInfo(symbol string) *marketInfo {
	mi := pdt.markets[symbol]
	if mi == nil {
		mi = newMarketInfo()
		pdt.markets[symbol] = mi
	}
	return mi
}

func (pdt *PDTrader) dispatchPumpSerachWork3(tArr []*Ticker) {
	wCh := make(chan []*Ticker)
	rCh := make(chan struct{})
	workersCnt := workerPoolSize
	for i := 0; i < workersCnt; i++ {
		go pdt.worker2(wCh, rCh)
	}
	sliceSize := len(tArr) / workersCnt
	lasti := workersCnt - 1
	for i := 0; i < workersCnt; i++ {
		if i == lasti {
			wCh <- tArr[sliceSize*i:]
		} else {
			offset := sliceSize * i
			wCh <- tArr[offset : offset+sliceSize]
		}
	}
	for {
		<-rCh
		workersCnt--
		if workersCnt == 0 {
			close(wCh)
			return
		}
	}
}

func (pdt *PDTrader) worker2(wCh chan []*Ticker, rCh chan struct{}) {
	for {
		tArr, ok := <-wCh
		if ok == false {
			return
		}
		for _, t := range tArr {
			pdt.marketsMutext.Lock()
			mi := pdt.marketInfo(t.Symbol)
			pdt.marketsMutext.Unlock()
			pdt.searchPump(t, mi)
		}
		rCh <- struct{}{}
	}
}

var count time.Duration = 0

func (pdt *PDTrader) searchPump(t *Ticker, mi *marketInfo) {
	// If buffer is not full just append
	if len(mi.buffer) < bufferSize {
		// Add ticket to buffer
		mi.buffer = append(mi.buffer, t)
	}
	// If trade is in progres just append
	if mi.pdCxt != nil {
		mi.lastNonZeroVelocity = 0
		mi.buffer = append(mi.buffer, t)
		mi.buffer = mi.buffer[1:]
	}

	// Compute derivatice
	prev := mi.buffer[len(mi.buffer)-1]
	dTime := goutils.Duration(t.Timestamp.Sub(prev.Timestamp)).UnixFloatNano()
	dPrice := t.Price - prev.Price
	velocity := dPrice / dTime
	pPrice := t.Price / prev.Price
	// fmt.Println("Price:", fmt.Sprintf("%.6f", t.Price),
	// 	", dTime:", fmt.Sprintf("%.3f", dTime),
	// 	", dPrice:", fmt.Sprintf("%.6f", dPrice),
	// 	", pPrice:", fmt.Sprintf("%.3f", pPrice),
	// 	", Velocity:", fmt.Sprintf("%.8f", velocity))
	// If velocity jumped order of magnitude open trade
	if (velocity/mi.lastNonZeroVelocity) > 150 && pPrice > 1.03 {
		//mi.pdCxt = newPdContext(t)
		//pdt.PumpDetected(mi.pdCxt)
		log.Println("\n\n")
		log.Println("Pump & dump trade triggered\n",
			"Id:", t.ID,
			", Sybol: ", t.Symbol,
			", Price:", fmt.Sprintf("%.6f", t.Price),
			", dTime:", fmt.Sprintf("%.3f", dTime),
			", dPrice:", fmt.Sprintf("%.6f", dPrice),
			", pPrice:", fmt.Sprintf("%.3f", pPrice),
			", Velocity:", fmt.Sprintf("%.8f", velocity))
		log.Println("\n\n")
	}
	if velocity > 0 {
		mi.lastNonZeroVelocity = velocity
	}
	// Add ticker to buffer and remove old
	if t != nil {
		mi.buffer = append(mi.buffer, t)
	}
	mi.buffer = mi.buffer[1:]
}

type workerItem struct {
	t  *Ticker
	mi *marketInfo
}
