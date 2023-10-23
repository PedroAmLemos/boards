package main

import "fmt"

var (
	createBoardSignal chan bool
	isBoard           bool = false
	boards                 = make(map[string]*Board)
	thisIP            string
	thisName          string
	people            = make(map[string]string)
)

func main() {
	createBoardSignal = make(chan bool)
	var fileName string
	thisName, fileName = getArgs()
	people = readFile(fileName)
	thisIP = people[thisName]
	fmt.Println("this ip", thisIP)
	delete(people, thisName)
	waitServerStart := make(chan bool)
	go startServer(thisIP, waitServerStart)
	<-waitServerStart
	go mainLoop(people, thisName)
	for range createBoardSignal {
		isBoard = true
		mainBoard := NewBoard("mainBoard")
		boards["mainBoard"] = mainBoard
		mainBoard.Start()
	}

}
