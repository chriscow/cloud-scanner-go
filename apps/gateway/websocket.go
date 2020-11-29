package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nsqio/go-nsq"

	"github.com/go-chi/render"
	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pingPeriod = 60 * time.Second
)

type wsMsgHandler struct {
	conn *websocket.Conn
}

func (h *wsMsgHandler) HandleMessage(msg *nsq.Message) error {

	return nil
}

func wsHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("[WSHandler] called")
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Instantiate a consumer that will subscribe to the provided channel.
	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(resultsTopic, "client-websocket", config)
	if err != nil {
		render.Render(w, r, ErrServerError("NewConsumer", err))
		return
	}

	// Set the Handler for messages received by this Consumer. Can be called multiple times.
	// See also AddConcurrentHandlers.
	h := &wsMsgHandler{conn: conn}
	consumer.AddHandler(h)

	// Use nsqlookupd to discover nsqd instances.
	// See also ConnectToNSQD, ConnectToNSQDs, ConnectToNSQLookupds.
	err = consumer.ConnectToNSQLookupd(os.Getenv("NSQ_LOOKUP") + ":4161")
	// err = consumer.ConnectToNSQD("localhost:4150")
	if err != nil {
		render.Render(w, r, ErrServerError("ConnectToNSQLookupd", err))
		return
	}

	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer func() {
			ticker.Stop()
			h.conn.Close()
		}()

		for {
			select {
			case <-ticker.C:
				h.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := h.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					log.Println("webosocket client closed", err)
				} else {
					log.Println("ReadMessage error:", err)
				}
				return
			}
		}
	}()
}
