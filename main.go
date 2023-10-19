package main

var createBoardSignal chan bool

func main() {
	createBoardSignal = make(chan bool)

	go mainLoop()
	for {
		select {
		case <-createBoardSignal:
			createBoard()
		}
	}

}

