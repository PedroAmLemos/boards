package main

import "fmt"

var createBoardSignal chan bool
var board bool = false

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
	for {
		select {
		case <-createBoardSignal:
			board = true
			createBoard()
		}
	}

}
