package main

import "fmt"

type BoardAction struct {
	Action    string
	BoardName string
}

func main() {
	createBoardSignal := make(chan BoardAction)
	thisName, fileName := getArgs()
	nodes := readFile(fileName, thisName)

	printStartScreen(nodes)

	activeBoard := false

	go mainLoop(nodes, createBoardSignal, &activeBoard)

	waitServerStart := make(chan bool)
	stopCheck := false
	go startServer(nodes, waitServerStart, &activeBoard, &stopCheck, createBoardSignal)
	<-waitServerStart

	for boardAction := range createBoardSignal {
		fmt.Printf("\n[board] %s\n> ", boardAction)
		switch boardAction.Action {
		case "changeOwner":
			fmt.Println()
			printHorizontalLine()
			fmt.Printf("[board] Starting board %s\n", boardAction.BoardName)
			activeBoard = true
			go checkBoardConnection(nodes, boardAction.BoardName, createBoardSignal, &activeBoard, &stopCheck)
			stopCheck = false
			nodes[boardAction.BoardName].board.Start(nodes, &activeBoard)
		case "newOwner":
			fmt.Println()
			printHorizontalLine()
			fmt.Printf("\nLost connection to %v\nThis will be the new owner\n", boardAction.BoardName)
			board := nodes[boardAction.BoardName].board
			if board != nil {
				board.name = "mainBoard"
				nodes["thisNode"].board = board
				nodes[boardAction.BoardName].board = nil
				multicast(nodes, fmt.Sprintf("newboardowner %v %v", boardAction.BoardName, nodes["thisNode"].name))
				activeBoard = true
				board.Start(nodes, &activeBoard)
				clients := nodes["thisNode"].board.connectedClients
				for name := range clients {
					unicast(nodes, name, fmt.Sprintf("boarddeleted %v", nodes["thisNode"].name))
				}
				activeBoard = false
				nodes["thisNode"].board = nil
				fmt.Printf("\n[board] Board %s closed\n", boardAction.BoardName)
				printHorizontalLine()
				fmt.Printf("\n> ")
			}
		case "new":
			fmt.Printf("\nCreating new board %s\n", boardAction.BoardName)
			activeBoard = true
			board := NewBoard("mainBoard")
			nodes["thisNode"].board = board
			board.Start(nodes, &activeBoard)
			fmt.Printf("\n[board] Disconnecting from board %s\n", boardAction.BoardName)
			clients := nodes["thisNode"].board.connectedClients
			for name := range clients {
				unicast(nodes, name, fmt.Sprintf("boarddeleted %v", nodes["thisNode"].name))
			}
			activeBoard = false
			nodes["thisNode"].board = nil
			fmt.Printf("\n[board] Board %s closed\n", boardAction.BoardName)
			printHorizontalLine()
			fmt.Printf("\n> ")
		case "connect":
			go checkBoardConnection(nodes, boardAction.BoardName, createBoardSignal, &activeBoard, &stopCheck)
			boardOwner := nodes[boardAction.BoardName]
			if boardOwner == nil {
				fmt.Printf("\n[board] Board %s does not exist\n", boardAction.BoardName)
				continue
			}
			fmt.Printf("\n[board] Connecting to board %s\n", boardAction.BoardName)
			response, err := unicast(nodes, boardOwner.name, "connecttoboard")
			if err != nil {
				fmt.Printf("\n[error] Error connecting to board %s: %s\n", boardAction.BoardName, err)
				continue
			}
			if string(response) == "false" {
				fmt.Printf("\n[board] %s does not have a created board\n", boardAction.BoardName)
				continue
			}
			activeBoard = true
			lines := parseLines(string(response))
			board := NewBoard(boardOwner.name)
			for _, line := range lines {
				board.AddLine(line)
			}
			nodes[board.name].board = board
			board.Start(nodes, &activeBoard)
			fmt.Printf("\n[board] Board %s deleted\n", boardAction.BoardName)
			// nodes[boardAction.BoardName].board = nil
			// unicast(nodes, boardOwner.name, fmt.Sprintf("clientdisconnected %v", nodes["thisNode"].name))
			stopCheck = true
			activeBoard = false
			unicast(nodes, boardOwner.name, fmt.Sprintf("clientdisconnected %v %v", boardOwner.name, nodes["thisNode"].name))
			printHorizontalLine()
			fmt.Printf("\n> ")
		}
	}
}
