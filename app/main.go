package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/db"
	"github.com/codecrafters-io/redis-starter-go/app/parser"
)

var _ = net.Listen
var _ = os.Exit

var SupportedCommands = map[string]bool{
	"ECHO":  true,
	"PING":  true,
	"SET":   true,
	"GET":   true,
	"RPUSH": true,
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	db := db.NewDb()
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

func handleConnection(conn net.Conn, db *db.Db) {
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

		output, err := RunCommand(value, db)
		if err != nil {
			serializedError := serializeOutput(err, true)
			conn.Write(serializedError)
			continue
		}
		outputSerialized := serializeOutput(output, false)
		if outputSerialized == nil {
			serializedError := serializeOutput(fmt.Errorf("unsupported protocol type"), true)
			conn.Write(serializedError)
			continue
		}
		conn.Write(outputSerialized)
	}

}

func RunCommand(input any, db *db.Db) (any, error) {
	arrAsAny, ok := input.([]any)
	if !ok || len(arrAsAny) == 0 {
		return "", fmt.Errorf("command must be an array of strings")
	}

	arr, err := AnyToString(arrAsAny)
	if err != nil {
		return "", err
	}

	commandName := arr[0]
	commandName = strings.ToUpper(commandName)

	command, err := command.NewCommand(commandName, db)
	if err != nil {
		return "", err
	}
	return command.ExecuteCommand(arr)

}
func serializeOutput(output any, isError bool) []byte {
	if output == "PONG" {
		return []byte(fmt.Sprintf("+%s\r\n", output))
	}

	if isError {
		return []byte(fmt.Sprintf("-%s\r\n", output))
	}

	switch v := output.(type) {
	case string:
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
	case int, int64, int32:
		return []byte(fmt.Sprintf(":%d\r\n", v))
	case []string:
		return []byte(fmt.Sprintf("*%d\r\n", len(v))) // Note: Real implementation needs to serialize each element.

	case nil:
		return []byte("$-1\r\n")

	default:
		return nil
	}
}

func AnyToString(input []any) ([]string, error) {
	var arrStr []string
	for _, v := range input {
		if str, ok := v.(string); ok {
			arrStr = append(arrStr, str)
		} else {
			return nil, fmt.Errorf("expected string argument")
		}
	}
	return arrStr, nil
}
