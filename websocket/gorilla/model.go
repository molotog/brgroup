package gorilla

import (
	"fmt"
	"strconv"
	"time"

	"go.brgroup.com/brgroup/websocket"
)

type Subscribe struct {
	Op string `json:"op"`
	ID string `json:"id"`
	Ch string `json:"ch"`
}

type Message struct {
	M      string `json:"m"`
	Symbol string `json:"symbol"`
	Data   struct {
		Ts  int64    `json:"ts"`
		Bid []string `json:"bid"`
		Ask []string `json:"ask"`
	} `json:"data"`
}

func (m Message) IsBBO() bool {
	return m.M == "bbo"
}

func (m Message) ToBestOrderBook() websocket.BestOrderBook {
	if len(m.Data.Ask) < 2 || len(m.Data.Bid) < 2 {
		return websocket.BestOrderBook{}
	}

	return websocket.BestOrderBook{
		Ask: websocket.Order{
			Amount: ValueFloat(m.Data.Ask[0]),
			Price:  ValueFloat(m.Data.Ask[1]),
		},
		Bid: websocket.Order{
			Amount: ValueFloat(m.Data.Bid[0]),
			Price:  ValueFloat(m.Data.Bid[1]),
		},
	}
}

func ValueFloat(value string) float64 {
	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return val
}

func NewBBOSubscribe(symbol string) Subscribe {
	return Subscribe{
		Op: "sub",
		ID: strconv.Itoa(int(time.Now().UnixMilli())),
		Ch: fmt.Sprintf("bbo:%s", symbol),
	}
}
