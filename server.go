package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	dinamicBuffer := make([]byte, 1024)
	n, err := conn.Read(dinamicBuffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return
	}

	msg := string(dinamicBuffer[:n])
	switch msg {
	case "PING":
		fmt.Printf("\n[log] Received PING, sending PONG back\n")
		conn.Write([]byte("PONG"))
	case "listBoards":
		fmt.Printf("\n[log] Received listBoards, sending %t back\n", board)
		if board {
			conn.Write([]byte("true"))
		} else {
			conn.Write([]byte("false"))
		}
	default:
		fmt.Printf("\n[log] Received unknown message: %s\n", msg)
	}
	fmt.Printf("> ")
}

func startServer(ip string, waitServerStart chan bool) {
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
			go handleConnection(conn)
		}
	}

	if retries == maxRetries {
		fmt.Println("Failed to start the server after multiple attempts. Exiting.")
		os.Exit(1)
	}
}
