package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

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

func readFile(fileName string) map[string]string {
	var lines []string
	hosts, _ := os.Open(fileName)
	defer hosts.Close()
	scanner := bufio.NewScanner(hosts)
	personMap := make(map[string]string)

	for scanner.Scan() {
		if len(scanner.Text()) == 0 {
			continue
		}
		lines = append(lines, scanner.Text())
	}

	for _, line := range lines {
		words := strings.Split(line, " ")
		personMap[words[1]] = words[0]
	}

	return personMap
}
