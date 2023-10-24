package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	raylib "github.com/gen2brain/raylib-go/raylib"
)

func handleConnection(conn net.Conn, people map[string]string, boards map[string]*Board, isBoard *bool, connectedClients map[string]string) {
	defer conn.Close()
	dinamicBuffer := make([]byte, 1024)
	n, err := conn.Read(dinamicBuffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return
	}

	msg := string(dinamicBuffer[:n])
	msgParts := strings.Split(msg, " ")
	name := msgParts[0]
	cmd := msgParts[1]
	switch cmd {
	case "PING":
		fmt.Printf("\n[log] Received PING, sending PONG back\n")
		conn.Write([]byte("PONG"))
	case "listBoards":
		fmt.Printf("\n[log] Received listBoards, sending %t back\n", *isBoard)
		if *isBoard {
			conn.Write([]byte("true"))
		} else {
			conn.Write([]byte("false"))
		}
	case "connectToBoard":
		fmt.Printf("\n[log] Received connectToBoard, sending the lines back\n")
		conn.Write([]byte(boards["mainBoard"].GetLines()))
		connectedClients[name] = people[name]
		fmt.Printf("\n[log] New client connected: %v\n> ", name)
	case "newLine":
		boardName := string(msgParts[2])
		fmt.Printf("\n[log] Received newLine from %v to %v, printing it into the board\n", name, boardName)
		conn.Write([]byte("Line created"))
		x1, err := strconv.ParseFloat(msgParts[3], 32)
		if err != nil {
			fmt.Println("[error] Error parsing float:", err.Error())
			return
		}
		y1, err := strconv.ParseFloat(msgParts[4], 32)
		if err != nil {
			fmt.Println("[error] Error parsing float:", err.Error())
			return
		}
		x2, err := strconv.ParseFloat(msgParts[5], 32)
		if err != nil {
			fmt.Println("[error] Error parsing float:", err.Error())
			return
		}
		y2, err := strconv.ParseFloat(msgParts[6], 32)
		if err != nil {
			fmt.Println("[error] Error parsing float:", err.Error())
			return
		}
		fmt.Printf("\n[log] New line at %.2f %.2f %.2f %.2f from %v\n> ", x1, y1, x2, y2, boardName)
		boards[boardName].AddLine(Line{
			Start: raylib.Vector2{X: float32(x1), Y: float32(y1)},
			End:   raylib.Vector2{X: float32(x2), Y: float32(y2)},
		})
		if boardName == "mainBoard" {
			fmt.Printf("\n[log] Sending to %v\n", connectedClients)
		}
	default:
		fmt.Printf("\n[log] Received unknown message: %s\n", msg)
		conn.Write([]byte("Received unknown message"))
	}
	fmt.Printf("> ")
}

func startServer(ip string, waitServerStart chan bool, people map[string]string, boards map[string]*Board, isBoard *bool, connectedClients map[string]string) {
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

		fmt.Printf("[log] Server started successfully on port: %s\n", ip)
		waitServerStart <- true

		defer ln.Close()
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("Error accepting connection:", err)
				continue
			}
			go handleConnection(conn, people, boards, isBoard, connectedClients)
		}
	}

	if retries == maxRetries {
		fmt.Println("Failed to start the server after multiple attempts. Exiting.")
		os.Exit(1)
	}
}
