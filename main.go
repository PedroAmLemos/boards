package main

import "fmt"

var (
	createBoardSignal chan bool
	isBoard           bool = false
	boards                 = make(map[string]*Board)
)

func main() {
	createBoardSignal = make(chan bool)

	name, fileName := getArgs()
	people := readFile(fileName)
	thisIP := people[name]
	fmt.Println("this ip", thisIP)
	delete(people, name)
	waitServerStart := make(chan bool)
	go startServer(thisIP, waitServerStart)
	<-waitServerStart
	go mainLoop(people, name)
	for range createBoardSignal {
		isBoard = true
		mainBoard := NewBoard("mainBoard")
		boards["mainBoard"] = mainBoard
		mainBoard.Start()
	}

}
