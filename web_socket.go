package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"sync"
)
var (
	wsClients      = make(map[*websocket.Conn]bool)
	wsClientsMux   sync.Mutex
)

func values(m map[string]*User) []*User {
	v := make([]*User, 0, len(m))
	for _, u := range m {
		v = append(v, u)
	}
	return v
}

func broadcast() {
	usersMux.Lock()
	data, _ := json.Marshal(values(users))
	usersMux.Unlock()

	wsClientsMux.Lock()
	defer wsClientsMux.Unlock()
	for conn := range wsClients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			conn.Close()
			delete(wsClients, conn)
		}
	}
}
