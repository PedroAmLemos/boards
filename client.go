package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	raylib "github.com/gen2brain/raylib-go/raylib"
)

func unicast(name, recipient, message string) ([]byte, error) {
	fmt.Printf("[log] Sending message to %s at %s: %s\n", name, recipient, message)
	conn, err := net.Dial("tcp", recipient)
	if err != nil {
		fmt.Println("[error] Error connecting to recipient:", err)
		return nil, err
	}
	defer conn.Close()
	msg := fmt.Sprintf("%v %v", name, message)
	_, err = conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("[error] Error sending message:", err)
		return nil, err
	}
	response := make([]byte, 1024)
	n, err := conn.Read(response)
	if err != nil {
		fmt.Println("[error] Error reading response:", err)
		return nil, err
	}
	return response[:n], nil
}

func multicast(name string, people map[string]string, message string) (map[string][]byte, error) {
	fmt.Printf("[log] Sending message to all peers: %s\n", message)
	responses := make(map[string][]byte)
	for name, ip := range people {
		conn, err := net.Dial("tcp", ip)
		if err != nil {
			fmt.Println("[error] Error connecting to recipient:", err)
		}
		defer func() {
			if conn != nil {
				conn.Close()
			}
		}()
		if conn != nil {
			msg := fmt.Sprintf("%v %v", name, message)
			_, err = conn.Write([]byte(msg))
		} else {
			continue
		}
		if err != nil {
			fmt.Println("[error] Error sending message:", err)
		}
		response := make([]byte, 1024)
		n, err := conn.Read(response)
		if err != nil {
			fmt.Println("[error] Error reading response:", err)
			return nil, err
		}
		responses[name] = response[:n]
	}
	return responses, nil
}

func mainLoop(people map[string]string, thisName string, boards map[string]*Board, isBoard *bool, createBoardSignal chan bool, connectedClients map[string]string) {
	for {
		cmd := readInput("> ")
		switch cmd {
		case "createBoard":
			createBoardSignal <- true
		case "unicast":
			sentTo := readInput("Enter the name of the person you want to send the message to: ")
			message := readInput("Enter the message: ")
			response, err := unicast(sentTo, people[sentTo], message)
			if err != nil {
				fmt.Println("[error] Error sending message:", err)
				continue
			}
			fmt.Printf("[log] Response from %s: %s\n", thisName, string(response))
		case "multicast":
			message := readInput("Enter the message: ")
			responses, err := multicast(thisName, people, message)
			if err != nil {
				fmt.Println("[error] Error sending message:", err)
				continue
			}
			fmt.Printf("[log] Responses from peers: %s\n", responses)

		case "listBoards":
			responses, err := multicast(thisName, people, "listBoards")
			if err != nil {
				fmt.Println("[error] Error sending message:", err)
				continue
			}
			for name, response := range responses {
				if string(response) == "true" {
					fmt.Printf("[log] %s has a board\n", name)
				}
			}
		case "listConnectedBoards":
			for name := range boards {
				fmt.Printf("[log] %s has a board\n", name)
			}

		case "listConnectedClients":
			for name := range connectedClients {
				fmt.Printf("[log] Client name: %s has a board\n", name)
			}

		case "connectToBoard":
			boardName := readInput("Enter the name of the person you want to connect to: ")
			response, err := unicast(thisName, people[boardName], "connectToBoard")
			if err != nil {
				fmt.Println("[error] Error sending message:", err)
				continue
			}
			fmt.Printf("[log] Response from %s: %q\n", boardName, string(response))
			lines := parseLine(string(response))
			newBoard := NewBoard(boardName)
			newBoard.lines = lines
			boards[boardName] = newBoard
			fmt.Printf("[log] Board %v added", boards[boardName].name)
			go newBoard.Start(thisName, people, isBoard, connectedClients)
		}

	}
}

func parseLine(lines string) []Line {
	var result []Line

	linesArr := strings.Split(strings.TrimSpace(lines), "\n")

	numLines, err := strconv.Atoi(linesArr[0])
	if err != nil {
		fmt.Println("[error] Error parsing number of lines:", err)
		return result
	}

	for i := 1; i <= numLines; i++ {
		lineArr := strings.Split(linesArr[i], " ")
		if len(lineArr) != 4 {
			fmt.Println("[error] Invalid line format:", linesArr[i])
			continue
		}
		x1, err := strconv.ParseFloat(lineArr[0], 64)
		if err != nil {
			fmt.Println("[error] Error parsing x1:", err)
			continue
		}
		y1, err := strconv.ParseFloat(lineArr[1], 64)
		if err != nil {
			fmt.Println("[error] Error parsing y1:", err)
			continue
		}
		x2, err := strconv.ParseFloat(lineArr[2], 64)
		if err != nil {
			fmt.Println("[error] Error parsing x2:", err)
			continue
		}
		y2, err := strconv.ParseFloat(lineArr[3], 64)
		if err != nil {
			fmt.Println("[error] Error parsing y2:", err)
			continue
		}
		result = append(result, Line{
			Start: raylib.Vector2{X: float32(x1), Y: float32(y1)},
			End:   raylib.Vector2{X: float32(x2), Y: float32(y2)},
		})
	}

	return result
}
