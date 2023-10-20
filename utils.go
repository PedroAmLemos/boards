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

func mainLoop() {
	for {
		cmd := readInput("> ")
		switch cmd {
		case "createBoard":
			createBoardSignal <- true
		case "newLine":
			coords := readInput("Enter the coordinates: ")
			newLine, err := parseCoords(coords)
			if err != nil {
				fmt.Println("Error parsing coordinates. Please enter valid numbers.")
				continue
			} else {
				CreateLine(*newLine)
			}
		}

	}
}
