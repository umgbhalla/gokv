package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/umgbhalla/gokv/pkg/client"
)

func main() {
	// Initialize the GoKV client
	gokv := client.New("http://localhost:8080")

	fmt.Println("GoKV CLI")
	fmt.Println("Commands: GET <key>, SET <key> <value> [ttl], DELETE <key>, EXIT")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		command := scanner.Text()
		if err := executeCommand(gokv, command); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}

func executeCommand(gokv *client.Client, command string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil
	}

	switch strings.ToUpper(parts[0]) {
	case "GET":
		if len(parts) != 2 {
			return fmt.Errorf("usage: GET <key>")
		}
		return executeGet(gokv, parts[1])
	case "SET":
		if len(parts) < 3 {
			return fmt.Errorf("usage: SET <key> <value> [ttl]")
		}
		ttl := time.Duration(0)
		if len(parts) == 4 {
			duration, err := time.ParseDuration(parts[3])
			if err != nil {
				return fmt.Errorf("invalid TTL: %v", err)
			}
			ttl = duration
		}
		return executeSet(gokv, parts[1], parts[2], ttl)
	case "DELETE":
		if len(parts) != 2 {
			return fmt.Errorf("usage: DELETE <key>")
		}
		return executeDelete(gokv, parts[1])
	case "EXIT":
		fmt.Println("Goodbye!")
		os.Exit(0)
	default:
		return fmt.Errorf("unknown command: %s", parts[0])
	}
	return nil
}

func executeGet(gokv *client.Client, key string) error {
	value, err := gokv.Get(key)
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", value)
	return nil
}

func executeSet(gokv *client.Client, key, value string, ttl time.Duration) error {
	err := gokv.Set(key, value, ttl)
	if err != nil {
		return err
	}
	fmt.Println("OK")
	return nil
}

func executeDelete(gokv *client.Client, key string) error {
	err := gokv.Delete(key)
	if err != nil {
		return err
	}
	fmt.Println("OK")
	return nil
}
