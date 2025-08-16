package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/parser"
)

var _ = net.Listen
var _ = os.Exit

type Db struct {
	dbMap map[interface{}]interface{}

	mu *sync.Mutex
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
		err := parser.Parse()
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client closed the connection:", conn.RemoteAddr())
				return
			}
			fmt.Println("Error reading from connection: ", err.Error())
			conn.Write([]byte("-Error invalid command: '" + "'\r\n"))
			continue
		}
		output, err := db.RunCommand(parser.CurrentCommand)
		if err != nil {
			conn.Write([]byte("-Error running command: '" + "'\r\n"))
		}
		conn.Write([]byte(output))
	}

}

func (db *Db) RunCommand(command *parser.Command) (string, error) {

	var output interface{}
	var ok bool
	switch command.CommandName {
	case "PING":
		output = "PONG"
	case "ECHO":
		output = command.Arguments[0]
	case "SET":
		db.mu.Lock()
		db.dbMap[command.Arguments[0]] = command.Arguments[1]
		db.mu.Unlock()
		output = "OK"
	case "GET":
		output, ok = db.dbMap[command.Arguments[0]]
		if !ok {
			output = "-Key does not exist"
			return serializeOutput(output, true), nil
		}
	}
	return serializeOutput(output, false), nil
}

func serializeOutput(output interface{}, isError bool) string {
	if isError {
		return fmt.Sprintf("-%s\r\n", output)
	}
	return fmt.Sprintf("+%s\r\n", output)
}
