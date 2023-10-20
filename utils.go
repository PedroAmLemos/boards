package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func readInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func mainLoop(people map[string]string, name string) {
	for {
		cmd := readInput("> ")
		switch cmd {
		case "createBoard":
			createBoardSignal <- true
		case "newLine":
			coords := readInput("Enter the coordinates: ")
			newLine, err := parseCoords(coords)
			if err != nil {
				fmt.Println("[error] Error parsing coordinates. Please enter valid numbers.")
				continue
			} else {
				createLine(*newLine)
			}
		case "unicast":
			sentTo := readInput("Enter the name of the person you want to send the message to: ")
			message := readInput("Enter the message: ")
			response, err := unicast(name, people[sentTo], message)
			if err != nil {
				fmt.Println("[error] Error sending message:", err)
				continue
			}
			// handle response
			fmt.Printf("[log] Response from %s: %s\n", name, string(response))
		case "multicast":
			message := readInput("Enter the message: ")
			responses, err := multicast(people, message)
			if err != nil {
				fmt.Println("[error] Error sending message:", err)
				continue
			}
			// handle responses
			fmt.Printf("[log] Responses from peers: %s\n", responses)

		case "listBoards":
			responses, err := multicast(people, "listBoards")
			if err != nil {
				fmt.Println("[error] Error sending message:", err)
				continue
			}
			for name, response := range responses {
				if string(response) == "true" {
					fmt.Printf("[log] %s has a board\n", name)
				}
			}
		}

	}
}

func readFile(fileName string) map[string]string {
	var lines []string
	hosts, err := os.Open(fileName)
	if err != nil {
		fmt.Println("[error] Error reading file:", err)
		os.Exit(1)
	}
	defer hosts.Close()
	scanner := bufio.NewScanner(hosts)
	personMap := make(map[string]string)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	for _, line := range lines {
		words := strings.Split(line, " ")
		personMap[words[1]] = words[0]
	}

	return personMap
}

func getArgs() (string, string) {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Please provide the arguments")
		os.Exit(1)
	}
	name := args[0]
	fileName := args[1]
	return name, fileName
}
