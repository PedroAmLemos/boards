package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	raylib "github.com/gen2brain/raylib-go/raylib"
)

func unicast(nodes map[string]*Node, receiver string, message string) ([]byte, error) {
	fmt.Println()
	printHorizontalLine()
	fmt.Printf("[log] Unicasting %s\n", message)
	node := nodes[receiver]
	if node == nil {
		return nil, fmt.Errorf("node not found")
	}

	conn, err := net.Dial("tcp", node.ip)
	fmt.Printf("[log] Connecting to %s\n", node.ip)
	if err != nil {
		return nil, fmt.Errorf("could not connect to %s: %v", node.name, err)
	}
	defer conn.Close()
	msg := []byte(fmt.Sprintf("%v %v", nodes["thisNode"].name, message))
	_, err = conn.Write(msg)
	if err != nil {
		return nil, fmt.Errorf("could not send message to %s: %v", node.name, err)
	}
	fmt.Printf("[log] Sent message to %s\n[log] Waiting for response...", node.name)
	response := make([]byte, 1024)
	n, err := conn.Read(response)
	if err != nil {
		return nil, fmt.Errorf("could not read response from %s: %v", node.name, err)
	}
	fmt.Printf("\n[log] Got response from %s\n", node.name)
	printHorizontalLine()
	return response[:n], nil
}

func multicast(nodes map[string]*Node, message string) map[string][]byte {
	printHorizontalLine()
	fmt.Printf("[log] Multicasting %s\n", message)
	responses := make(map[string][]byte)

	for _, node := range nodes {
		if !node.isThisNode {
			conn, err := net.Dial("tcp", node.ip)
			fmt.Printf("[log] Connecting to %s\n", node.ip)
			if err != nil {
				fmt.Printf("[error] Could not connect to %s: %v\nContinuing...\n", node.name, err)
				continue
			}
			defer conn.Close()
			msg := []byte(fmt.Sprintf("%v %v", node.name, message))
			_, err = conn.Write(msg)
			if err != nil {
				fmt.Printf("[error] Could not send message to %s: %v\nContinuing...\n", node.name, err)
				continue
			}
			fmt.Printf("[log] Sent message to %s\n[log] Waiting for response...", node.name)
			response := make([]byte, 1024)

			n, err := conn.Read(response)
			if err != nil {
				fmt.Printf("[error] Could not read response from %s: %v\nContinuing...\n", node.name, err)
				continue
			}
			fmt.Printf("\n[log] Got response from %s\n", node.name)
			responses[node.name] = response[:n]
		}
	}
	printHorizontalLine()
	return responses
}

func mainLoop(nodes map[string]*Node, createBoardSignal chan BoardAction, activeBoard *bool) {
	for {
		command := strings.ToLower(readInput("> "))
		switch command {
		case "exit":
			os.Exit(0)
		case "createboard":
			if *activeBoard {
				fmt.Println("You are already connected to a board")
				continue
			}
			createBoardSignal <- BoardAction{Action: "new", BoardName: "mainBoard"}
		case "connecttoboard":
			if *activeBoard {
				fmt.Println("You are already connected to a board")
				continue
			}
			boardName := readInput("Enter the board name: ")
			createBoardSignal <- BoardAction{Action: "connect", BoardName: boardName}
		case "newline":
			nodes["thisNode"].board.AddLine(Line{
				Start: raylib.Vector2{X: 0, Y: 0},
				End:   raylib.Vector2{X: 100, Y: 100},
			})
		case "listcreatedboards":
		case "lscb":
			responses := multicast(nodes, "listcreatedboards")
			count := 0
			printHorizontalLine()
			for name, response := range responses {
				if string(response) == "true" {
					fmt.Printf("\n%v[log] %s has a created board%v\n", RedColor, name, ResetColor)
					count++
				}
			}
			if count == 0 {
				if nodes["thisNode"].board == nil {
					printRed("[log] no boards found")
				} else {
					printRed("[log] no boards found, only the current board is created")
				}
			}
			fmt.Println()
			printHorizontalLine()
		case "help":
			printHorizontalLine()
			printCommands()
			printHorizontalLine()
		default:
			fmt.Println("command not found")
		}
	}
}
