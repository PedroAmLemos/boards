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
	go startServer(nodes, waitServerStart)
	<-waitServerStart

	for boardAction := range createBoardSignal {
		fmt.Printf("\n[board] %s\n> ", boardAction)
		switch boardAction.Action {
		case "new":
			fmt.Printf("\nCreating new board %s\n", boardAction.BoardName)
			activeBoard = true
			board := NewBoard("mainBoard")
			nodes["thisNode"].board = board
			board.Start(nodes, &activeBoard)
		case "connect":
			// go checkConnection
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
		}
	}
}
