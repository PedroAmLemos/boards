package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var (
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan Line)
)

var boardExists = false

func startServer(ip string) {
	const maxRetries = 3
	retries := 0

	for retries < maxRetries {
		ln, err := net.Listen("tcp", ip)
		if err != nil {
			fmt.Printf("Error starting server: %s. Attempt %d/%d\n", err, retries+1, maxRetries)
			retries++
			time.Sleep(2 * time.Second)
			continue
		}

		defer ln.Close()
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("Error accepting connection:", err)
				continue
			}
			go handleConnection(conn)
		}
	}

	if retries == maxRetries {
		fmt.Println("Failed to start the server after multiple attempts. Exiting.")
		os.Exit(1)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Connection established with:", conn.RemoteAddr())
}

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
	x1, err1 := strconv.ParseFloat(parts[0], 64)
	y1, err2 := strconv.ParseFloat(parts[1], 64)
	x2, err3 := strconv.ParseFloat(parts[2], 64)
	y2, err4 := strconv.ParseFloat(parts[3], 64)

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

func readInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func createBoard() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Starting server on http://localhost:%d", listener.Addr().(*net.TCPAddr).Port)
	err = http.Serve(listener, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func mainLoop(people map[string]string, name string) {
	for {
		command := readInput("")
		switch command {
		case "exit":
			fmt.Println("Exiting...")
			os.Exit(0)
		case "createBoard":
			fmt.Println("Creating board...")
			if boardExists {
				fmt.Println("Board already exists")
				continue
			}
			go createBoard()
			fmt.Println("Board created")
			boardExists = true
		case "createLine":
			if !boardExists {
				fmt.Println("Board does not exist")
				continue
			}
			createLine(readInput("Enter coordinates: "))
		default:
			fmt.Println("Invalid command")
		}
	}
}

func main() {
	name, fileName := getArgs()
	people := readFile(fileName)
	thisIP := people[name]
	delete(people, name)
	fmt.Println("People: ", people)
	fmt.Println("This IP: ", thisIP)
	go mainLoop(people, name)

	startServer(thisIP)
}
