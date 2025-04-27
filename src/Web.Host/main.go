package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

)

// ----------------------------------------------------------------
// DBモデル定義
// ----------------------------------------------------------------

// Answer は数独の各セルの解答状態を表します。
type Answer struct {
	ID         uint   `gorm:"primary_key"`
	CellIndex  int    `gorm:"column:cell_index"`  // 1～81
	CellNumber int    `gorm:"column:cell_number"` // 解答（例：0=空 or 数字）
	SessionID  string `gorm:"column:session_id"`  // ルーム（セッション）ID
}

// Room は各セッション（ルーム）の情報を保持します。
type Room struct {
	SessionID string    `gorm:"primary_key;column:session_id"`
	CreatedAt time.Time
	Answers   []Answer  `gorm:"foreignkey:SessionID;association_foreignkey:SessionID"`
}

// ----------------------------------------------------------------
// インメモリルーム管理（WebSocket接続管理用）
// ----------------------------------------------------------------

type RoomHub struct {
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

func NewRoomHub() *RoomHub {
	return &RoomHub{clients: make(map[*websocket.Conn]bool)}
}

func (hub *RoomHub) Broadcast(message []byte) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	for client := range hub.clients {
		if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Println("Broadcast WriteMessage error:", err)
			client.Close()
			delete(hub.clients, client)
		}
	}
}

var (
	rooms   = make(map[string]*RoomHub)
	roomsMu sync.Mutex

	db *gorm.DB
)

// ----------------------------------------------------------------
// WebSocket サーバ実装
// ----------------------------------------------------------------

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
    sessionID := r.URL.Query().Get("sessionID")
    if sessionID == "" {
        http.Error(w, "sessionID is required", http.StatusBadRequest)
        return
    }

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("WebSocket Upgrade error:", err)
        return
    }
    defer conn.Close()

    // --- ここでは初期盤面をすぐ送らない！ --- 

    roomsMu.Lock()
    hub, exists := rooms[sessionID]
    if !exists {
        hub = NewRoomHub()
        rooms[sessionID] = hub
    }
    roomsMu.Unlock()

    hub.mu.Lock()
    hub.clients[conn] = true
    hub.mu.Unlock()

    // ここからクライアントごとにメッセージ受信ループ
    for {
        messageType, msg, err := conn.ReadMessage()
        if err != nil {
            log.Println("ReadMessage warning:", err)
            break
        }

        if messageType == websocket.TextMessage {
            var header struct {
                Type string `json:"type"`
            }
            if err := json.Unmarshal(msg, &header); err != nil {
                log.Println("JSON unmarshal error:", err)
                continue
            }

            switch header.Type {
            case "getInitialBoard":
                // 🔥 ここで、その時点で最新のRoom情報をDBから読む！
                var room Room
                if err := db.Preload("Answers").Where("session_id = ?", sessionID).First(&room).Error; err != nil {
                    log.Println("DB fetch error:", err)
                    continue
                }

                initialMessage, err := json.Marshal(room.Answers)
                if err != nil {
                    log.Println("JSON marshal error:", err)
                    continue
                }

                if err := conn.WriteMessage(websocket.TextMessage, initialMessage); err != nil {
                    log.Println("Write initial board error:", err)
                } else {
                    log.Println("Sent initial board to this connection!")
                }

            case "updateCell":
                var update struct {
                    CellIndex  int `json:"cellIndex"`
                    CellNumber int `json:"cellNumber"`
                }
                if err := json.Unmarshal(msg, &update); err != nil {
                    log.Println("JSON unmarshal error:", err)
                    continue
                }

                if err := db.Model(&Answer{}).
                    Where("session_id = ? AND cell_index = ?", sessionID, update.CellIndex).
                    Update("cell_number", update.CellNumber).Error; err != nil {
                    log.Println("DB update error:", err)
                } else {
                    // 全員に反映
                    hub.Broadcast(msg)
                }
            default:
                log.Println("Unknown message type:", header.Type)
            }
        }
    }

    // 切断処理
    roomsMu.Lock()
    if hub, ok := rooms[sessionID]; ok {
        hub.mu.Lock()
        delete(hub.clients, conn)
        hub.mu.Unlock()
    }
    roomsMu.Unlock()
}

func main() {
	var err error
	// DSN（例: "root:root@tcp(localhost:3306)/sudoku_db?parseTime=true"）は環境に合わせて変更してください
	db, err = gorm.Open("mysql", "user:passpass@tcp(localhost:3307)/sudoku-db?parseTime=true")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Room, Answer テーブルの自動マイグレーション
	db.AutoMigrate(&Room{}, &Answer{})

	// WebSocket ハンドラをセットアップ
	http.HandleFunc("/ws", wsHandler)
	log.Println("WebSocket server running on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal("ListenAndServe error:", err)
	}
}
