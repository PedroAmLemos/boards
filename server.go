package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func handleConnection(nodes map[string]*Node, conn net.Conn, activeBoard *bool, stopChan chan struct{}) {
	defer conn.Close()
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from connection: ", err)
		return
	}
	msg := string(buffer[:n])
	fmt.Printf("[log] Received message: %s\n> ", msg)
	msgParts := strings.Split(msg, " ")
	name := msgParts[0]
	cmd := msgParts[1]
	switch cmd {
	case "listcreatedboards":
		fmt.Println()
		printHorizontalLine()
		fmt.Printf("\n[log] %s is asking if this node has a board created\n> ", name)
		if nodes["thisNode"].board != nil {
			conn.Write([]byte("true"))
		} else {
			conn.Write([]byte("false"))
		}
		printHorizontalLine()
		fmt.Printf("\n> ")

	case "connecttoboard":
		fmt.Println()
		printHorizontalLine()
		fmt.Printf("\n[log] %s is asking to connect to this node's board\n> ", name)
		if nodes["thisNode"].board == nil {
			conn.Write([]byte("false"))
			return
		}
		lines := nodes["thisNode"].board.GetLines()
		conn.Write([]byte(lines))
		nodes["thisNode"].board.connectedClients[name] = struct{}{}
		printHorizontalLine()
		fmt.Printf("\n> ")
	case "boarddeleted":
		fmt.Println()
		printHorizontalLine()
		fmt.Printf("\n[log] %s has deleted it's board\n> ", name)
		*activeBoard = false
		printHorizontalLine()
		fmt.Printf("\n> ")
	case "clientdisconnected":
		fmt.Println()
		printHorizontalLine()
		fmt.Printf("\n[log] %s has disconnected from the board\n> ", name)
		delete(nodes["thisNode"].board.connectedClients, name)
		conn.Write([]byte("true"))
		printHorizontalLine()
		fmt.Printf("\n> ")
	case "newboardowner":
		fmt.Println()
		printHorizontalLine()
		stopChan <- struct{}{}
		// 		oldOwner := msgParts[2]
		// 		newOwner := msgParts[3]
		// 		if nodes["thisNode"].board != nil {
		// 			if nodes["thisNode"].board.name == oldOwner {
		// 				fmt.Printf("\n[log] %s is the new owner of the board\n> ", newOwner)
		// 				nodes["thisNode"].board.name = "mainBoard"
		// 				nodes[newOwner].board = nodes["thisNode"].board
		// 				nodes[oldOwner].board = nil
		// 				// *activeBoard = false
		// 				printHorizontalLine()
		// 				fmt.Printf("\n> ")
		// 			}
		// 		}
	case "updateline":
		fmt.Println()
		printHorizontalLine()
		fmt.Printf("\n[log] %s is updating a line\n", name)
		boardName := msgParts[2]
		line := parseLine(msgParts[3:])
		if boardName == "mainBoard" {
			nodes["thisNode"].board.UpdateLine(line)
			for client := range nodes["thisNode"].board.connectedClients {
				if client != name {
					unicast(nodes, client, fmt.Sprintf("updateline %s %s", nodes["thisNode"].name, line.String()))
				}
			}
		} else {
			nodes[boardName].board.UpdateLine(line)
		}
		conn.Write([]byte("true"))
		fmt.Println()
		printHorizontalLine()
		fmt.Printf("\n> ")

	case "newline":
		fmt.Println()
		printHorizontalLine()
		fmt.Printf("\n[log] %s is creating a new line\n", name)
		boardName := msgParts[2]
		line := parseLine(msgParts[3:])
		if boardName == "mainBoard" {
			nodes["thisNode"].board.AddLine(line)
			for client := range nodes["thisNode"].board.connectedClients {
				if client != name {
					unicast(nodes, client, fmt.Sprintf("newline %s %s", nodes["thisNode"].name, line.String()))
				}
			}
		} else {
			nodes[boardName].board.AddLine(line)
		}
		conn.Write([]byte("true"))
		fmt.Println()
		printHorizontalLine()
		fmt.Printf("\n> ")

	default:
		fmt.Printf("\n[log] Received unknown message: %s\n", msg)
		conn.Write([]byte("Received unknown message"))
		fmt.Printf("\n> ")
	}
}

func startServer(nodes map[string]*Node, waitServerStart chan bool, activeBoard *bool, stopChan chan struct{}) {
	const maxRetries = 3
	retries := 0
	ip := nodes["thisNode"].ip

	for retries < maxRetries {
		ln, err := net.Listen("tcp", ip)
		if err != nil {
			fmt.Printf("Error starting server: %s. Attempt %d/%d\n", err, retries+1, maxRetries)
			retries++
			time.Sleep(2 * time.Second)
			continue
		}

		fmt.Printf("\n[log] Server started successfully on port: %s\n> ", ip)
		waitServerStart <- true

		defer ln.Close()
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("Error accepting connection:", err)
				continue
			}
			fmt.Printf("[log] Accepted connection from %s\n> ", conn.RemoteAddr())
			go handleConnection(nodes, conn, activeBoard, stopChan)
		}
	}

	if retries == maxRetries {
		fmt.Println("Failed to start the server after multiple attempts. Exiting.")
		os.Exit(1)
	}
}
