package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	if !connectBot() {
		os.Exit(0)
	}

	go func() {
		sig := <-sigs
		fmt.Printf("\nReceived signal: %s. Performing cleanup...\n", sig)
		disconnectBot()
		os.Exit(0)
	}()

	// Gin web server
	r := gin.Default()
	r.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		wsClientsMux.Lock()
		wsClients[conn] = true
		wsClientsMux.Unlock()

		usersMux.Lock()
		conn.WriteJSON(values(users))
		usersMux.Unlock()
	})

	fmt.Println("Overlay server running at :8080")
	r.Run(":8080")
	
}
