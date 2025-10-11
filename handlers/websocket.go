package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/bstempelj/memory-kana/storage"
	"github.com/gorilla/websocket"
)

var ErrGameOver = errors.New("game over")

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: replace this placeholder with real origin checking
		return true
	},
}

type Game struct {
	pairs     []Pair
	startTime int64
	endTime   int64
	duration  time.Duration
}

type Timestamp struct {
	Timestamp int64 `json:"timestamp"`
}

type Pair struct {
	Kana      string `json:"kana"`
	Romaji    string `json:"romaji"`
	Timestamp int64  `json:"timestamp"`
}

type GameMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// shared across websocket connections
type WebSocketHandler struct {
	db *sql.DB
}

func NewWebSocketHandler(db *sql.DB) *WebSocketHandler {
	return &WebSocketHandler{
		db: db,
	}
}

// TODO: send error messages to client
func (ws *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var game Game

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("error upgrading", "err", err)
		return
	}
	slog.Debug("websocket connection upgraded")

	defer func() {
		if err := conn.Close(); err != nil {
			slog.Error("closing websocket connection", "err", err)
			return
		}
		slog.Debug("websocket connection closed")
	}()

	for {
		_, connMsg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				slog.Error("unexpected websocket close error", "err", err)
				return
			}
			slog.Error("websocket message read", "err", err)
			return
		}

		var msg GameMessage
		if err := json.Unmarshal(connMsg, &msg); err != nil {
			slog.Error("error unmarshaling message", "err", err)
			continue
		}
		slog.Debug("received message",
			"type", string(msg.Type),
			"data", string(msg.Data))

		if err := handleGameMessage(&game, msg); err != nil {
			if errors.Is(err, ErrGameOver) {
				break
			}
			slog.Error("handling game message", "err", err)
			continue
		}

	}

	start := time.Unix(game.startTime, 0)
	end := time.Unix(game.endTime, 0)
	game.duration = end.Sub(start)

	slog.Debug("game stats",
		"start", game.startTime,
		"end", game.endTime,
		"duration", game.duration)

	playerName, err := storage.InsertPlayerDuration(ws.db, game.duration)
	if err != nil {
		slog.Error("storing game duration", "err", err)
		return
	}

	clientMsg := map[string]any{
		"type": "gameover",
		"data": map[string]string{
			"redirect": "/scoreboard?p=" + playerName,
		},
	}

	if err := conn.WriteJSON(clientMsg); err != nil {
		slog.Error("sending gameover message to client", "err", err)
	}
}

func handleGameMessage(game *Game, msg GameMessage) error {
	switch msg.Type {
	case "start":
		var startTimestamp Timestamp
		if err := json.Unmarshal(msg.Data, &startTimestamp); err != nil {
			return err
		}

		game.startTime = startTimestamp.Timestamp
		slog.Debug("game start", "time", game.startTime)

	case "end":
		var endTimestamp Timestamp
		if err := json.Unmarshal(msg.Data, &endTimestamp); err != nil {
			return err
		}

		game.endTime = endTimestamp.Timestamp
		slog.Debug("game over", "time", game.endTime)

		return ErrGameOver

	case "pair":
		var pair Pair
		if err := json.Unmarshal(msg.Data, &pair); err != nil {
			return err
		}
		game.pairs = append(game.pairs, pair)

		slog.Debug(
			"received pair message",
			"kana", string(pair.Kana),
			"romaji", string(pair.Romaji),
			"timestamp", pair.Timestamp)
	}
	return nil
}
