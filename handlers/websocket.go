package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/bstempelj/memory-kana/storage"
	"github.com/gorilla/websocket"
)

// var hiragana = map[string]string{
// 	"あ": "a", "い": "i", "う": "u", "え": "e", "お": "o",
// 	"か": "ka", "き": "ki", "く": "ku", "け": "ke", "こ": "ko",
// 	"さ": "sa", "し": "shi", "す": "su", "せ": "se", "そ": "so",
// 	"た": "ta", "ち": "chi", "つ": "tsu", "て": "te", "と": "to",
// 	"な": "na", "に": "ni", "ぬ": "nu", "ね": "ne", "の": "no",
// 	"は": "ha", "ひ": "hi", "ふ": "fu", "へ": "he", "ほ": "ho",
// 	"ま": "ma", "み": "mi", "む": "mu", "め": "me", "も": "mo",
// 	"や": "ya", "ゆ": "yu", "よ": "yo",
// 	"ら": "ra", "り": "ri", "る": "ru", "れ": "re", "ろ": "ro",
// 	"わ": "wa", "を": "wo",
// 	"ん": "n"
// }

// var katakana = map[string]string{
// 	"ア": "a", "イ": "i", "ウ": "u", "エ": "e", "オ": "o",
// 	"カ": "ka", "キ": "ki", "ク": "ku", "ケ": "ke", "コ": "ko",
// 	"サ": "sa", "シ": "shi", "ス": "su", "セ": "se", "ソ": "so",
// 	"タ": "ta", "チ": "chi", "ツ": "tsu", "テ": "te", "ト": "to",
// 	"ナ": "na", "ニ": "ni", "ヌ": "nu", "ネ": "ne", "ノ": "no",
// 	"ハ": "ha", "ヒ": "hi", "フ": "fu", "ヘ": "he", "ホ": "ho",
// 	"マ": "ma", "ミ": "mi", "ム": "mu", "メ": "me", "モ": "mo",
// 	"ヤ": "ya", "ユ": "yu", "ヨ": "yo",
// 	"ラ": "ra", "リ": "ri", "ル": "ru", "レ": "re", "ロ": "ro",
// 	"ワ": "wa", "ヲ": "wo",
// 	"ン": "n"
// }

var hiragana = [46]string{
	"あ", "い", "う", "え", "お",
	"か", "き", "く", "け", "こ",
	"さ", "し", "す", "せ", "そ",
	"た", "ち", "つ", "て", "と",
	"な", "に", "ぬ", "ね", "の",
	"は", "ひ", "ふ", "へ", "ほ",
	"ま", "み", "む", "め", "も",
	"や", "ゆ", "よ",
	"ら", "り", "る", "れ", "ろ",
	"わ", "を",
	"ん",
}

var katakana = [46]string{
	"ア", "イ", "ウ", "エ", "オ",
	"カ", "キ", "ク", "ケ", "コ",
	"サ", "シ", "ス", "セ", "ソ",
	"タ", "チ", "ツ", "テ", "ト",
	"ナ", "ニ", "ヌ", "ネ", "ノ",
	"ハ", "ヒ", "フ", "ヘ", "ホ",
	"マ", "ミ", "ム", "メ", "モ",
	"ヤ", "ユ", "ヨ",
	"ラ", "リ", "ル", "レ", "ロ",
	"ワ", "ヲ",
	"ン",
}

var ErrGameOver = errors.New("game over")

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: replace this placeholder with real origin checking
		return true
	},
}

type Game struct {
	startTime int64
	endTime   int64
	duration  time.Duration
}


type GameInitData struct {
	Kana string `json:"kana"`
}

type GameStartData struct {
	Timestamp int64 `json:"timestamp"`
}

type GameEndData struct {
	Timestamp int64 `json:"timestamp"`
}

type GamePairData struct {
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
	case "init":
		var data GameInitData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			return err
		}

		var kana []string
		switch data.Kana {
		case "hiragana":
			kana = hiragana[:]
		case "katakana":
			kana = katakana[:]
		}
		slog.Debug("original", "kana", kana)
		fisherYatesShuffle(kana)
		slog.Debug("shuffled", "kana", kana)

		slog.Debug("game init", "kana", data.Kana, "tiles", kana)

	case "start":
		var data GameStartData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			return err
		}

		game.startTime = data.Timestamp
		slog.Debug("game start", "time", game.startTime)

	case "end":
		var data GameEndData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			return err
		}

		game.endTime = data.Timestamp
		slog.Debug("game over", "time", game.endTime)

		return ErrGameOver

	case "pair":
		var data GamePairData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			return err
		}

		slog.Debug(
			"received pair message",
			"kana", string(data.Kana),
			"romaji", string(data.Romaji),
			"timestamp", data.Timestamp)
	}
	return nil
}

func fisherYatesShuffle(kana []string) {
	for i := len(kana) - 1; i >= 1; i-- {
		j := rand.IntN(i + 1)
		kana[i], kana[j] = kana[j], kana[i]
	}
}
