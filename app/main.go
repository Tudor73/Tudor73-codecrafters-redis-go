package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/parser"
)

var _ = net.Listen
var _ = os.Exit

type MapValue struct {
	Value    any
	SetAt    time.Time
	ExpireAt time.Time
}

type Db struct {
	dbMap map[any]MapValue

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
		dbMap: make(map[any]MapValue),
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

func (db *Db) RunCommand(command any) (any, error) {
	arr, ok := command.([]any)
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

	var output any
	var err error
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
		output, err = db.RunSetCommand(arr)
		if err != nil {
			return nil, err
		}
	case "GET":
		if len(arr) != 2 {
			return nil, fmt.Errorf("invalid number of arguments for GET command %d", len(arr))
		}
		key := arr[1]
		val, ok := db.dbMap[key]
		if !ok {
			output = -1
			return output, nil
		}
		if time.Now().Compare(val.ExpireAt) == 1 {
			fmt.Println("key expired")
			output = -1
			return output, nil
		}
		output = val.Value
	}
	return output, nil
}

func (db *Db) RunSetCommand(command []any) (any, error) {
	if len(command) < 3 {
		return nil, fmt.Errorf("invalid number of arguments for SET command %d", len(command))
	}

	newValue := MapValue{
		Value: command[2],
	}
	if len(command) == 5 {
		flag, ok := command[3].(string)
		if !ok {
			return nil, fmt.Errorf("unsupported type for option %s", flag)
		}
		flag = strings.ToUpper(flag)
		if flag != "PX" {
			return nil, fmt.Errorf("unsupported option %s", flag)
		}

		durationAsString, _ := command[4].(string)

		duration, err := strconv.Atoi(durationAsString)
		if err != nil {
			return nil, fmt.Errorf("invalid data type for option, expected number %s", flag)
		}
		newValue.SetAt = time.Now()
		newValue.ExpireAt = time.Now().Add(time.Millisecond * time.Duration(duration))
	}

	db.mu.Lock()
	db.dbMap[command[1]] = newValue
	db.mu.Unlock()
	return "OK", nil

}

func serializeOutput(output any, isError bool) string {
	if isError {
		return fmt.Sprintf("-%s\r\n", output)
	}
	return fmt.Sprintf("+%s\r\n", output)
}
