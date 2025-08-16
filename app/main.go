package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/parser"
)

var _ = net.Listen
var _ = os.Exit

type Db struct {
	dbMap map[interface{}]interface{}

	mu *sync.Mutex
}

var SupportedCommands = map[string]bool{
	"ECHO": true,
	"PING": true,
	"SET":  true,
	"GET":  true,
}

func NewDb() *Db {
	return &Db{
		dbMap: make(map[interface{}]interface{}),
		mu:    &sync.Mutex{},
	}

}

func main() {
	fmt.Println("Logs from your program will appear here!")

	db := NewDb()
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		go handleConnection(conn, db)
	}
}

func handleConnection(conn net.Conn, db *Db) {
	defer conn.Close()

	fmt.Println("Handling Connection", conn.RemoteAddr())
	reader := bufio.NewReader(conn)
	parser := parser.NewParser(reader)
	for {
		value, err := parser.Parse()
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client closed the connection:", conn.RemoteAddr())
				return
			}
			fmt.Println("Error reading from connection: ", err.Error())
			conn.Write([]byte("-Error invalid command: '" + "'\r\n"))
			continue
		}

		output, err := db.RunCommand(value)
		if err != nil {
			serializedError := serializeOutput(err, true)
			conn.Write([]byte(serializedError))
		}
		serializedOutput := serializeOutput(output, false)
		conn.Write([]byte(serializedOutput))
	}

}

func (db *Db) RunCommand(command interface{}) (interface{}, error) {
	arr, ok := command.([]interface{})
	if !ok || len(arr) == 0 {
		return nil, fmt.Errorf("command must be an array")
	}

	commandName, ok := arr[0].(string)
	if !ok {
		return nil, fmt.Errorf("first element must be a string")
	}
	commandName = strings.ToUpper(commandName)
	if _, supported := SupportedCommands[commandName]; !supported {
		return nil, fmt.Errorf("unsupported command: %s", commandName)
	}

	var output interface{}
	switch commandName {
	case "PING":
		output = "PONG"
	case "ECHO":
		argument, ok := arr[1].(string)
		if !ok {
			return nil, fmt.Errorf("argument for echo command must be a string")
		}
		output = argument
	case "SET":
		if len(arr) != 3 {
			return nil, fmt.Errorf("invalid number of arguments for SET command %s", len(arr))
		}
		key := arr[1]
		value := arr[2]
		db.mu.Lock()
		db.dbMap[key] = value
		db.mu.Unlock()
		output = "OK"
	case "GET":
		if len(arr) != 2 {
			return nil, fmt.Errorf("invalid number of arguments for GET command %s", len(arr))
		}
		key := arr[1]
		val, ok := db.dbMap[key]
		if !ok {
			output = -1
			return output, nil
		}
		output = val
	}
	return output, nil
}

func serializeOutput(output interface{}, isError bool) string {
	if isError {
		return fmt.Sprintf("-%s\r\n", output)
	}
	return fmt.Sprintf("+%s\r\n", output)
}
