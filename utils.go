package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	Unicast                     = "unicast"
	Multicast                   = "multicast"
	MulticastDelay              = "multicast-delay"
	Heartbeat                   = "heartbeat"
	DefaultIntervalForHeartbeat = 1
	ResetColor                  = "\033[0m"
	RedColor                    = "\033[31m"
	GreenColor                  = "\033[32m"
	YellowColor                 = "\033[33m"
	BlueColor                   = "\033[34m"
	MagentaColor                = "\033[35m"
	CyanColor                   = "\033[36m"
)

type Node struct {
	name            string
	ip              string
	isAlive         bool
	isThisNode      bool
	expectedTimeout float64
	lastHeartbeat   time.Time
	board           *Board
}

func NewNode(name string, ip string, isAlive bool, isThisNode bool, expectedTimeout float64) *Node {
	return &Node{name, ip, isAlive, isThisNode, expectedTimeout, time.Time{}, nil}
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

func readFile(fileName string, thisName string) map[string]*Node {
	nodes := make(map[string]*Node)
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		if len(parts) == 2 {
			ip := parts[0]
			name := parts[1]
			if name == thisName {
				nodes["thisNode"] = NewNode(name, ip, true, true, DefaultIntervalForHeartbeat)
			} else {
				nodes[name] = NewNode(name, ip, true, false, DefaultIntervalForHeartbeat)
			}
		}
	}
	return nodes
}

func printHorizontalLine() {
	fmt.Println(CyanColor + "------------------------------------------------" + ResetColor)
}

func printStartScreen(nodes map[string]*Node) {
	printHorizontalLine()
	fmt.Println(BlueColor + "Welcome " + nodes["thisNode"].name + ResetColor)
	fmt.Printf(YellowColor+"Your IP is: %v\n"+ResetColor, nodes["thisNode"].ip)
	fmt.Println(MagentaColor + "People found in the file: " + ResetColor)
	for _, node := range nodes {
		if !node.isThisNode {
			fmt.Printf("%s: %s\n", node.name, node.ip)
		}
	}
	fmt.Println("Type " + RedColor + "'help'" + ResetColor + " to see the list of commands")
	printHorizontalLine()
}

func printCommands() {
	printHorizontalLine()
	fmt.Println("Type '" + RedColor + "exit" + ResetColor + "' to exit")
	fmt.Println("Type '" + RedColor + "list" + ResetColor + "' to list all people")
	fmt.Println("Type '" + RedColor + "multicast" + ResetColor + "' to send a message to all people")
	fmt.Println("Type '" + RedColor + "unicast" + ResetColor + "' to send a message to a specific person")
	fmt.Println("Type '" + RedColor + "clear" + ResetColor + "' to clear the screen")
	fmt.Println("Type '" + RedColor + "help" + ResetColor + "' to see this list of commands")
	// fmt.Println("Type '" + RedColor + "get-delay" + ResetColor + "' to get the current delay")
	// fmt.Println("Type '" + RedColor + "set-delay" + ResetColor + "' to set the current delay")
	fmt.Println("Type '" + RedColor + "status" + ResetColor + "' to get the nodes status")
	// fmt.Println("Type '" + RedColor + "set-multicast-delay" + ResetColor + "' to set the multicast delay")
	// fmt.Println("Type '" + RedColor + "set-unicast-delay" + ResetColor + "' to set the unicast delay")
	fmt.Println("Type '" + RedColor + "createboard" + ResetColor + "' to create a new board")
	fmt.Println("Type '" + RedColor + "connecttoboard" + ResetColor + "' to connect to a board")
	// listcreatedboards or lscb to list the created boards
	fmt.Println(
		"Type '" + RedColor + "listcreatedboards" + ResetColor + "' or '" + RedColor + "lscb" + ResetColor + "'to list the created boards",
	)

	printHorizontalLine()
}

func readInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func centerText(text string, width int) string {
	padding := width - len(text)
	leftPadding := padding / 2
	rightPadding := padding - leftPadding
	return strings.Repeat("=", leftPadding) + " " + text + " " + strings.Repeat("=", rightPadding)
}

func printRed(content string) {
	fmt.Printf("\n%s%s%s\n", RedColor, content, ResetColor)
}
