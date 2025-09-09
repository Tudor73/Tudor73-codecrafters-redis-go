package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/commands"
	"github.com/codecrafters-io/redis-starter-go/app/db"
	"github.com/codecrafters-io/redis-starter-go/app/eventloop"
	"github.com/codecrafters-io/redis-starter-go/app/parser"
)

var _ = net.Listen
var _ = os.Exit

var SupportedCommands = map[string]bool{
	"ECHO":   true,
	"PING":   true,
	"SET":    true,
	"GET":    true,
	"RPUSH":  true,
	"LRANGE": true,
	"LPUSH":  true,
	"LLEN":   true,
	"LPOP":   true,
	"BLPOP":  true,
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	db := db.NewDb()
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	eventLoop := eventloop.NewEventLoop()
	go eventLoop.Run()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		go handleConnection(conn, db, eventLoop)
	}
}

func handleConnection(conn net.Conn, db *db.Db, queue *eventloop.EventLoop) {
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

		command, err := RunCommand(value, db, queue)
		if err != nil {
			serializedError := commands.SerializeOutput(err, true)
			conn.Write(serializedError)
			continue
		}
		resultChan := command.GetResponseChan()
		result := <-resultChan
		conn.Write(result)

	}

}

func RunCommand(input any, db *db.Db, queue *eventloop.EventLoop) (commands.Command, error) {
	arrAsAny, ok := input.([]any)
	if !ok || len(arrAsAny) == 0 {
		return nil, fmt.Errorf("command must be an array of strings")
	}

	arr, err := AnyToString(arrAsAny)
	if err != nil {
		return nil, err
	}

	commandName := arr[0]
	commandName = strings.ToUpper(commandName)

	command, err := commands.NewCommand(commandName, db, arr)
	if err != nil {
		return nil, err
	}
	queue.Tasks <- command
	return command, nil

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
