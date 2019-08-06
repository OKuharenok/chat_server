package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"time"
)

var DSN = "root:1234@tcp(mysql:3306)/golang?charset=utf8"
var db *sql.DB

// docker run -p 3306:3306 -v $(PWD):/docker-entrypoint-initdb.d -e MYSQL_ROOT_PASSWORD=1234 -e MYSQL_DATABASE=golang -d mysql

type User struct {
	UserId    string `json:"user,omitempty"`
	UserName  string `json:"username,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

type Chat struct {
	ChatId    string   `json:"chat,omitempty"`
	ChatName  string   `json:"name,omitempty"`
	Users     []string `json:"users,omitempty"`
	CreatedAt string   `json:"created_at,omitempty"`
	LastDate  string   `json:"-"`
}

type Message struct {
	MessageId int    `json:"message_id,omitempty"`
	ChatId    string `json:"chat"`
	UserId    string `json:"author"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at,omitempty"`
}

func createUser(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if body != nil {
		defer r.Body.Close()
	}
	if err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var usr User
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&usr)
	if err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	usr.CreatedAt = time.Now().String()[:19]
	res, err := db.Exec("INSERT INTO `users`(username, created_at) VALUES (?, ?)", usr.UserName, usr.CreatedAt)
	if err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	id, _ := res.LastInsertId()
	resp := make(map[string]int64)
	resp["id"] = id
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func createChat(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if body != nil {
		defer r.Body.Close()
	}
	if err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var chat Chat
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&chat)
	if err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	chat.CreatedAt = time.Now().String()[:19]
	res, err := db.Exec("INSERT INTO `chats`(chat_name, created_at) VALUES (?, ?)", chat.ChatName, chat.CreatedAt)
	if err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	id, _ := res.LastInsertId()
	for _, val := range chat.Users {
		userId, err := strconv.Atoi(val)
		if err != nil {
			logrus.Errorln(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, err = db.Exec("INSERT INTO chats_users(chat_id, user_id) VALUES (?, ?)", id, userId)
		if err != nil {
			logrus.Errorln(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	resp := make(map[string]int64)
	resp["id"] = id
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func createMessage(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if body != nil {
		defer r.Body.Close()
	}
	if err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var msg Message
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&msg)
	if err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	msg.CreatedAt = time.Now().String()[:19]
	chId, err := strconv.Atoi(msg.ChatId)
	usId, err := strconv.Atoi(msg.UserId)
	if err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	res, err := db.Exec("INSERT INTO `messages`(chat_id, user_id, text, created_at) VALUES (?, ?, ?, ?)", chId, usId, msg.Text, msg.CreatedAt)
	if err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	id, _ := res.LastInsertId()
	resp := make(map[string]int64)
	resp["id"] = id
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func getChats(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if body != nil {
		defer r.Body.Close()
	}
	if err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var usr User
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&usr)
	if err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	usId, err := strconv.Atoi(usr.UserId)
	if err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	rows, err := db.Query("SELECT chat_id FROM `chats_users` WHERE user_id = ?", usId)
	defer rows.Close()
	var chats []Chat
	for rows.Next() {
		var chat Chat
		var chId int
		rows.Scan(&chId)
		chat.ChatId = strconv.Itoa(chId)
		mrows, err := db.Query("SELECT created_at FROM `messages` WHERE chat_id = ? ORDER BY created_at DESC", chId)
		if err != nil {
			logrus.Errorln(err)
			w.WriteHeader(http.StatusInternalServerError)
			mrows.Close()
			return
		}
		mrows.Next()
		mrows.Scan(&chat.LastDate)
		mrows.Close()
		chrow := db.QueryRow("SELECT chat_name, created_at FROM `chats` WHERE chat_id = ?", chId)
		chrow.Scan(&chat.ChatName, &chat.CreatedAt)
		usrows, err := db.Query("SELECT user_id FROM `chats_users` WHERE chat_id = ?", chId)
		if err != nil {
			logrus.Errorln(err)
			w.WriteHeader(http.StatusInternalServerError)
			usrows.Close()
			return
		}
		for usrows.Next() {
			var uid int
			usrows.Scan(&uid)
			chat.Users = append(chat.Users, strconv.Itoa(uid))
		}
		usrows.Close()
		chats = append(chats, chat)
	}
	sort.Slice(chats, func(i, j int) bool {
		return chats[i].LastDate > chats[j].LastDate
	})
	err = json.NewEncoder(w).Encode(chats)
	if err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func getMessages(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if body != nil {
		defer r.Body.Close()
	}
	if err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var chat Chat
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&chat)
	if err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	chId, err := strconv.Atoi(chat.ChatId)
	if err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	rows, err := db.Query("SELECT message_id, chat_id, user_id, text, created_at FROM messages WHERE chat_id = ? ORDER BY created_at ASC", chId)
	defer rows.Close()
	if err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var msgs []Message
	for rows.Next() {
		var msg Message
		rows.Scan(&msg.MessageId, &msg.ChatId, &msg.UserId, &msg.Text, &msg.CreatedAt)
		msgs = append(msgs, msg)
	}
	err = json.NewEncoder(w).Encode(msgs)
	if err != nil {
		logrus.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func main() {
	var err error
	db, err = sql.Open("mysql", DSN)
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	router := mux.NewRouter()
	router.HandleFunc("/users/add", createUser).Methods("POST").Headers("Content-Type", "application/json")
	router.HandleFunc("/chats/add", createChat).Methods("POST").Headers("Content-Type", "application/json")
	router.HandleFunc("/messages/add", createMessage).Methods("POST").Headers("Content-Type", "application/json")
	router.HandleFunc("/chats/get", getChats).Methods("POST").Headers("Content-Type", "application/json")
	router.HandleFunc("/messages/get", getMessages).Methods("POST").Headers("Content-Type", "application/json")
	logrus.Infoln("starting server at :9000")
	err = http.ListenAndServe(":9000", router)
	if err != nil {
		logrus.Errorln(err)
	}
}
