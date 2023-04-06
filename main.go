package main

import (
	"os"

	"github.com/joho/godotenv"

	"go.brgroup.com/brgroup/config"
	"go.brgroup.com/brgroup/logger"
	"go.brgroup.com/brgroup/websocket"
	"go.brgroup.com/brgroup/websocket/gorilla"
)

func main() {
	log := logger.NewLogger()

	conf, err := loadConfig()
	if err != nil {
		log.Error("environment load failed", err)
	}

	log.Info("conf", map[string]interface{}{
		"conf": conf,
	})

	client := gorilla.NewClient(log, conf)

	err = client.Connection()
	if err != nil {
		log.Error("connection failed", err)
		return
	}
	defer client.Disconnect()

	done := make(chan websocket.BestOrderBook)
	client.ReadMessagesFromChannel(done)

	symbols := []string{
		"USDT_BTC",
		"BTC_USDT",
	}
	for _, symbol := range symbols {
		err := client.SubscribeToChannel(symbol)
		if err != nil {
			log.Error("subscribe to channel failed", err, map[string]interface{}{
				"symbol": symbol,
			})
		}
	}

	go client.WriteMessagesToChannel()

	for book := range done {
		log.Info("book", map[string]interface{}{
			"book": book,
		})
	}
}

func loadConfig() (config.Ascendex, error) {
	err := godotenv.Load()
	if err != nil {
		return config.Ascendex{}, err
	}

	return config.Ascendex{
		Host:      os.Getenv("ASCENDEX_HOST"),
		Scheme:    os.Getenv("ASCENDEX_SCHEME"),
		Path:      os.Getenv("ASCENDEX_PATH"),
		APIKey:    os.Getenv("ASCENDEX_APIKEY"),
		APISecret: os.Getenv("ASCENDEX_APISECRET"),
	}, nil
}
