package main

import (
	"io/ioutil"
	"github.com/gorilla/websocket"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var (
	clients = make(map[float64]*websocket.Conn, 1)
	mu sync.Mutex
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func homePage(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile("index.html")
	if err == nil {
		w.Write(data)
	}
}

func reader(msg_prefix string, client_id float64, conn *websocket.Conn) {
	for {
		messageType, p, err := conn.ReadMessage()
		if err == nil {
			for id, client := range clients {
				err = client.WriteMessage(messageType, []byte(msg_prefix + string(p)))
				if err != nil {
					mu.Lock()
					delete(clients, id)
					mu.Unlock()
					if client_id == id {
						return
					}
				}
			}
		} else {
			return
		}
	}
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err == nil {
		client_id := rand.Float64()
		client_id_str := strconv.FormatFloat(client_id, 'f', 10, 64) 
		msg := []byte("User " + client_id_str + " joined!")
		err = ws.WriteMessage(1, msg)
		if err == nil {
			for id, client := range clients {
				err = client.WriteMessage(1, msg)
				if err != nil {
					mu.Lock()
					delete(clients, id)
					mu.Unlock()
				}
			}
			mu.Lock()
			clients[client_id] = ws
			mu.Unlock()
			reader(client_id_str + ": ", client_id, ws)
		}
	}
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", wsEndpoint)
	panic(http.ListenAndServe(":80", nil))
}