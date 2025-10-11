package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/bstempelj/memory-kana/storage"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: replace this placeholder with real origin checking
		return true
	},
}

type Game struct {
	pairs          []Pair
	startTimestamp int64
	endTimestamp   int64
}

type Timestamp struct {
	Timestamp int64 `json:"timestamp"`
}

type Pair struct {
	Kana      string `json:"kana"`
	Romaji    string `json:"romaji"`
	Timestamp int64  `json:"timestamp"`
}

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type WebSocketHandler struct {
	mu       sync.Mutex
	db       *sql.DB
	connPool map[*websocket.Conn]*Game
}

func NewWebSocketHandler(db *sql.DB) WebSocketHandler {
	return WebSocketHandler{
		db:       db,
		connPool: make(map[*websocket.Conn]*Game),
	}
}

func (self WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("error upgrading", "err", err)
		return
	}
	slog.Info("websocket connection upgraded")

	self.mu.Lock()
	self.connPool[conn] = &Game{}
	self.mu.Unlock()

	slog.Info("websocket connection registered")

	defer func() {
		self.mu.Lock()
		delete(self.connPool, conn)
		self.mu.Unlock()

		slog.Info("websocket connection unregistered")

		if err := conn.Close(); err != nil {
			slog.Error("error closing websocket connection", "err", err)
			return
		}
		slog.Info("websocket connection closed")
	}()

	for {
		_, connMsg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("error reading message", "err", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(connMsg, &msg); err != nil {
			slog.Error("error unmarshaling message", "err", err)
			continue
		}
		slog.Info("received message", "type", string(msg.Type), "data", string(msg.Data))

		switch msg.Type {
		case "start":
			var startTimestamp Timestamp
			if err := json.Unmarshal(msg.Data, &startTimestamp); err != nil {
				slog.Error("error unmarshaling start message", "err", err)
				continue
			}

			self.mu.Lock()
			self.connPool[conn].startTimestamp = startTimestamp.Timestamp
			self.mu.Unlock()
		case "end":
			var endTimestamp Timestamp
			if err := json.Unmarshal(msg.Data, &endTimestamp); err != nil {
				slog.Error("error unmarshaling end message", "err", err)
				continue
			}

			self.mu.Lock()
			self.connPool[conn].endTimestamp = endTimestamp.Timestamp
			self.mu.Unlock()

			self.mu.Lock()
			start := time.Unix(self.connPool[conn].startTimestamp, 0)
			end := time.Unix(self.connPool[conn].endTimestamp, 0)
			self.mu.Unlock()

			duration := end.Sub(start)
			slog.Info("game over", "duration(seconds)", duration.Seconds())

			playerName, err := storage.InsertPlayerDuration(self.db, duration)
			if err != nil {
				// TODO: redirect to error page
				log.Fatal(err)
			}

			conn.WriteJSON(map[string]any{
				"type": "gameover",
				"data": map[string]string{
					"redirect": "/scoreboard?p=" + playerName,
				},
			})
			break
		case "pair":
			var pair Pair
			if err := json.Unmarshal(msg.Data, &pair); err != nil {
				slog.Error("error unmarshaling pair message", "err", err)
				continue
			}
			slog.Info(
				"received pair message",
				"kana", string(pair.Kana),
				"romaji", string(pair.Romaji),
				"timestamp", pair.Timestamp)

			self.mu.Lock()
			self.connPool[conn].pairs = append(self.connPool[conn].pairs, pair)
			self.mu.Unlock()
		default:
			fmt.Println(string(msg.Data))
		}
	}
}
