package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"github.com/gin-gonic/gin"
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

	r := gin.Default()
	r.GET("/ws", func(c *gin.Context) {
		_, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
	})

	fmt.Println("Overlay server running at :8080")
	r.Run(":8080")
	
}
