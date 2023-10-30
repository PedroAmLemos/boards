package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func handleConnection(nodes map[string]*Node, conn net.Conn, activeBoard *bool, changedOwner *bool, createBoardSignal chan BoardAction) {
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
		// send list of connected clients to the new client and notify the others
		clients := nodes["thisNode"].board.connectedClients
		clientsList := ""
		for client := range clients {
			clientsList += fmt.Sprintf("%s ", client)
		}
		unicast(nodes, name, fmt.Sprintf("connectedclients %s", clientsList))
		for client := range clients {
			if client != name {
				unicast(nodes, client, fmt.Sprintf("newclient %s", name))
			}
		}
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
		oldOwner := msgParts[2]
		newOwner := msgParts[3]
		fmt.Printf("\n[log] Change of ownership: %v to %v\n", oldOwner, newOwner)
		if nodes[oldOwner].board != nil {
			fmt.Printf("\n[log] Stoping check for board %s\n", oldOwner)
			*activeBoard = false
			*changedOwner = true
			time.Sleep(2 * time.Second)
			// nodes[oldOwner].board = nil
			board := nodes[oldOwner].board
			board.name = newOwner
			nodes[newOwner].board = board
			nodes[oldOwner].board = nil
			createBoardSignal <- BoardAction{BoardName: newOwner, Action: "changeOwner"}
			fmt.Printf("\n[log] %s is the new owner of the board %s\n", newOwner, oldOwner)
			printHorizontalLine()
			fmt.Printf("\n> ")
		}
		fmt.Printf("\n[log] %s is the new owner of the board %s\n> ", newOwner, oldOwner)
		conn.Write([]byte("true"))
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
	case "connectedclients":
		fmt.Println()
		printHorizontalLine()
		fmt.Printf("\n[log] %s is sending the list of connected clients\n", name)
		clients := msgParts[2:]
		for _, client := range clients {
			nodes[name].board.connectedClients[client] = struct{}{}
		}
		printHorizontalLine()
		fmt.Printf("\n> ")
	case "newclient":
		fmt.Println()
		printHorizontalLine()
		fmt.Printf("\n[log] %s is notifying that a new client has connected\n", name)
		nodes[name].board.connectedClients[msgParts[2]] = struct{}{}
		printHorizontalLine()
		fmt.Printf("\n> ")
	default:
		fmt.Printf("\n[log] Received unknown message: %s\n", msg)
		conn.Write([]byte("Received unknown message"))
		fmt.Printf("\n> ")
	}
}

func startServer(nodes map[string]*Node, waitServerStart chan bool, activeBoard *bool, changedOwner *bool, createBoardSignal chan BoardAction) {
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
			go handleConnection(nodes, conn, activeBoard, changedOwner, createBoardSignal)
		}
	}

	if retries == maxRetries {
		fmt.Println("Failed to start the server after multiple attempts. Exiting.")
		os.Exit(1)
	}
}
