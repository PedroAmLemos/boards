package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

var (
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan Line)
)

type Line struct {
	X1 float64 `json:"x1"`
	Y1 float64 `json:"y1"`
	X2 float64 `json:"x2"`
	Y2 float64 `json:"y2"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()

	clients[ws] = true

	for {
		var msg map[string]interface{}
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		switch msg["action"] {
		case "create":
			log.Printf("Created line: %v", msg["line"])
		case "update":
			log.Printf("Updated line: %v", msg["line"])
		}
	}
}

func handleMessages() {
	for {
		line := <-broadcast
		for client := range clients {
			err := client.WriteJSON(line)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func createLine(cmd string) {
	parts := strings.Fields(cmd)
	if len(parts) < 5 || strings.ToLower(parts[0]) != "createline" {
		log.Println("Invalid command")
		return
	}

	x1, err1 := strconv.ParseFloat(parts[1], 64)
	y1, err2 := strconv.ParseFloat(parts[2], 64)
	x2, err3 := strconv.ParseFloat(parts[3], 64)
	y2, err4 := strconv.ParseFloat(parts[4], 64)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		log.Println("Invalid coordinates")
		return
	}

	broadcast <- Line{x1, y1, x2, y2}
}

func listenForCommands() {
	reader := bufio.NewReader(os.Stdin)
	for {
		cmd, _ := reader.ReadString('\n')
		createLine(strings.TrimSpace(cmd))
	}
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/ws", handleConnections)

	go handleMessages()
	go listenForCommands()

	port := "8080"
	log.Printf("Starting server on http://localhost:%s", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
