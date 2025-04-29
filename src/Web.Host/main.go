package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
    "math/rand"
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

// problem は数独の各セルの問題状態を表します。
type Problem struct {
    ID         uint   `gorm:"primary_key"`
    CellIndex  int    `gorm:"column:cell_index"`
    CellNumber int    `gorm:"column:cell_number"` // 0なら空白
    SessionID  string `gorm:"column:session_id"`
}

// Room は各セッション（ルーム）の情報を保持します。
type Room struct {
	SessionID string    `gorm:"primary_key;column:session_id"`
	CreatedAt time.Time
	Answers   []Answer  `gorm:"foreignkey:SessionID;association_foreignkey:SessionID"`
    Problems  []Problem `gorm:"foreignkey:SessionID;association_foreignkey:SessionID"`

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

// createProblem は数独の問題を生成します。
func createProblem(answers []answer) []Problem {
    const holeCount = 40 // 空白にするマスの数（調整可能）
    rand.Seed(time.Now().UnixNano())

    // 1〜81からholeCount個ランダムに穴を開けるマスを選ぶ
    indices := rand.Perm(81)[:holeCount]
    holes := make(map[int]bool)
    for _, idx := range indices {
        holes[idx+1] = true // 1始まりに合わせる
    }

    var problems []Problem
    for _, ans := range answers {
        cellNumber := ans.CellNumber
        if holes[ans.CellIndex] {
            cellNumber = 0 // 穴をあけるマスは 0 にする
        }
        problems = append(problems, Problem{
            CellIndex:  ans.CellIndex,
            CellNumber: cellNumber,
            SessionID:  ans.SessionID,
        })
    }
    return problems
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
                // ⭐ ここでRoomを探す（ProblemもPreload）
                var room Room
                log.Println("Room loaded:", room.SessionID, "Problems count:", len(room.Problems))
                if err := db.Preload("Problems").Where("session_id = ?", sessionID).First(&room).Error; err != nil {
                    log.Println("Room not found, generating new board...")

                    // 🔥 完成盤面を作る
                    generatedAnswers := createAnswer()

                    // 🔥 問題盤面を作る
                    generatedProblems := createProblem(generatedAnswers)

                    // 🔥 Roomを作って保存
                    var answers []Answer
                    for _, cell := range generatedAnswers {
                        answers = append(answers, Answer{
                            CellIndex:  cell.CellIndex,
                            CellNumber: cell.CellNumber,
                            SessionID:  sessionID,
                        })
                    }

                    var problems []Problem
                    for _, cell := range generatedProblems {
                        problems = append(problems, Problem{
                            CellIndex:  cell.CellIndex,
                            CellNumber: cell.CellNumber,
                            SessionID:  sessionID,
                        })
                    }

                    room = Room{
                        SessionID: sessionID,
                        CreatedAt: time.Now(),
                        Answers:   answers,
                        Problems:  problems,
                    }

                    if err := db.Create(&room).Error; err != nil {
                        log.Println("DB room creation error:", err)
                        continue
                    }
                }

                // ⭐ ここでは「出題盤面（Problem）」を返す！
                initialMessage, err := json.Marshal(room.Problems)
                if err != nil {
                    log.Println("JSON marshal error:", err)
                    continue
                }

                if err := conn.WriteMessage(websocket.TextMessage, initialMessage); err != nil {
                    log.Println("Write initial board error:", err)
                } else {
                    log.Println("Sent initial board to this connection!")
                    log.Println("Room loaded:", room.SessionID, "Problems count:", len(room.Problems))
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
	db.AutoMigrate(&Room{}, &Answer{}, &Problem{})

	// WebSocket ハンドラをセットアップ
	http.HandleFunc("/ws", wsHandler)
	log.Println("WebSocket server running on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal("ListenAndServe error:", err)
	}
}
