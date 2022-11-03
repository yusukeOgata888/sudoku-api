package main

import (
	"fmt"
	"log"
	"net/http"

	"example.com/Web.Host/utils"
	"github.com/gorilla/mux"
    "github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// User User構造体
type User struct {
    ID        int
    FirstName string
    LastName  string
}


func main() {
    router := mux.NewRouter().StrictSlash(true)
    router.HandleFunc("/", home)
    router.HandleFunc("/users", insertSolution)
    // router.HandleFunc("/users", findSolution)
    http.ListenAndServe(":8080", router)
}


func home(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello World")
}

func findAllUsers(w http.ResponseWriter, r *http.Request) {
    // DB接続
    db := utils.GetConnection()
    
    defer db.Close()
    var userList []User
    db.Find(&userList)

    // 共通化した処理を使う
    utils.RespondWithJSON(w, http.StatusOK, userList)
}

func insertSolution(w http.ResponseWriter, r *http.Request){
    
    w.Header().Set("Access-Control-Allow-Headers", "*")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set( "Access-Control-Allow-Methods","GET, POST, PUT, DELETE, OPTIONS" )
    var solutions []solutions
   
   // DB接続
   db, err := gorm.Open("mysql", "root:pass@unix(/var/lib/mysql/mysql.sock)/sudoku?charset=utf8&parseTime=True&loc=Local")
   // 接続に失敗したらエラーログを出して終了する
   if err != nil {
       log.Fatalf("DB connection failed %v", err)
   }
   db.LogMode(true)
   db.Delete(&solutions)

   solutions = createSolution()
    // DBにINSERTする
    for i := range solutions {
    db.Create(solutions[i])
    }
    defer db.Close()

    utils.RespondWithJSON(w, http.StatusOK, solutions)
}

func findSolution(w http.ResponseWriter, r *http.Request){
    // DB接続
    db, err := gorm.Open("mysql", "root:pass@unix(/var/lib/mysql/mysql.sock)/sudoku?charset=utf8&parseTime=True&loc=Local")
    // 接続に失敗したらエラーログを出して終了する
    if err != nil {
        log.Fatalf("DB connection failed %v", err)
    }
    db.LogMode(true)
    defer db.Close()

    var solutions solutions
    db.Find(&solutions)

    // 共通化した処理を使う
    //utils.RespondWithJSON(w, http.StatusOK, db)
}

func findByID(w http.ResponseWriter, r *http.Request) {

    id, err := utils.GetID(r)
    if err != nil {
        utils.RespondWithError(w, http.StatusBadRequest, "Invalid parameter")
        return
    }

    // DB接続
    db := utils.GetConnection()
    defer db.Close()

    var user User
    db.Where("id = ?", id).Find(&user)

    // 共通化した処理を使う
    utils.RespondWithJSON(w, http.StatusOK, user)
}
