package gorilla

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"

	gorillawebsocket "github.com/gorilla/websocket"

	"go.brgroup.com/brgroup/config"
	"go.brgroup.com/brgroup/logger"
	"go.brgroup.com/brgroup/websocket"
)

type Broadcaster interface {
	Broadcast()
}

type client struct {
	log  logger.Logger
	conf config.Ascendex
	conn *gorillawebsocket.Conn
}

func NewClient(
	log logger.Logger,
	conf config.Ascendex,
) websocket.APIClient {
	return &client{
		log:  log,
		conf: conf,
	}
}

func (c *client) Connection() error {
	u := url.URL{
		Scheme: c.conf.Scheme,
		Host:   c.conf.Host,
		Path:   c.conf.Path,
	}

	c.log.Info("connecting...", map[string]interface{}{
		"to": u.String(),
	})

	ts := strconv.Itoa(int(time.Now().UTC().UnixMilli()))
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(u.String(), map[string][]string{
		"x-auth-key":       {c.conf.APIKey},
		"x-auth-signature": {c.signature(strings.Join([]string{ts, "+", u.Path}, ""), c.conf.APISecret)},
		"x-auth-timestamp": {ts},
	})
	if err != nil {
		c.log.Error("dial failed", err, map[string]interface{}{
			"url": u.String(),
		})
		return err
	}
	c.conn = conn

	return nil
}

func (c *client) Disconnect() {
	c.log.Info("Disconnect")
	defer c.conn.Close()
}

func (c *client) SubscribeToChannel(symbol string) error {
	symbol = strings.ReplaceAll(symbol, "_", "/")

	sub, err := json.Marshal(NewBBOSubscribe(symbol))
	if err != nil {
		return err
	}

	err = c.conn.WriteMessage(gorillawebsocket.TextMessage, sub)
	if err != nil {
		c.log.Error("subscribe failed", err, map[string]interface{}{
			"symbol": symbol,
		})
		return err
	}

	c.log.Info("subscribed to channel",
		map[string]interface{}{
			"symbol": symbol,
		},
	)
	return nil
}

func (c *client) ReadMessagesFromChannel(ch chan<- websocket.BestOrderBook) {
	c.log.Info("ReadMessagesFromChannel")

	go func() {
		defer close(ch)
		for {
			messateType, message, err := c.conn.ReadMessage()
			if err != nil {
				c.log.Error("read:", err)
				return
			}

			var msg Message
			err = json.Unmarshal(message, &msg)
			if err != nil {
				c.log.Error("unmarshal message failed", err, map[string]interface{}{
					"message": string(message),
				})
				return
			}

			c.log.Info("received", map[string]interface{}{
				"message":     string(message),
				"messateType": messateType,
			})

			if msg.IsBBO() {
				ch <- msg.ToBestOrderBook()
			}

		}
	}()
}

func (c *client) WriteMessagesToChannel() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			c.log.Info("ping", map[string]interface{}{
				"message": t.String(),
			})

			err := c.conn.WriteControl(gorillawebsocket.PingMessage, []byte(`{ "op": "ping" }`), time.Now().UTC().Add(time.Second))
			if err != nil {
				c.log.Error("ping failed", err)
				return
			}
		}
	}
}

func (c *client) signature(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
