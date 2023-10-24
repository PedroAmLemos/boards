package main

import "fmt"

func main() {
	var (
		createBoardSignal = make(chan bool)
		boards            = make(map[string]*Board)
		connectedClients  = make(map[string]string)
		fileName          string
		isBoard           = false
	)
	thisName, fileName := getArgs()
	people := readFile(fileName)
	thisIP := people[thisName]
	fmt.Println("this ip", thisIP)
	delete(people, thisName)
	waitServerStart := make(chan bool)
	go startServer(thisIP, waitServerStart, people, boards, &isBoard, connectedClients, thisName)
	<-waitServerStart
	go mainLoop(people, thisName, boards, &isBoard, createBoardSignal, connectedClients)
	for range createBoardSignal {
		isBoard = true
		mainBoard := NewBoard("mainBoard")
		boards["mainBoard"] = mainBoard
		mainBoard.Start(thisName, people, &isBoard, connectedClients)
	}

}
