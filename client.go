package main

import (
	"fmt"
	"net"
)

func unicast(name, recipient, message string) ([]byte, error) {
	fmt.Printf("[log] Sending message to %s at %s: %s\n", name, recipient, message)
	conn, err := net.Dial("tcp", recipient)
	if err != nil {
		fmt.Println("[error] Error connecting to recipient:", err)
		return nil, err
	}
	defer conn.Close()
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("[error] Error sending message:", err)
		return nil, err
	}
	response := make([]byte, 1024)
	_, err = conn.Read(response)
	if err != nil {
		fmt.Println("[error] Error reading response:", err)
		return nil, err
	}
	return response, nil
}

func multicast(people map[string]string, message string) (map[string][]byte, error) {
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
			_, err = conn.Write([]byte(message))
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
