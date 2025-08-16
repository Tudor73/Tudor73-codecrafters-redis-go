package parser

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

type Token struct {
	Lexeme   string
	DataType string
	Value    interface{}
}

type Parser struct {
	Source         *bufio.Reader
	Start          int
	Current        int
	Err            error
	Tokens         []Token
	CurrentCommand *Command
}

var SupportedCommands = map[string]bool{
	"ECHO": true,
	"PING": true,
	"SET":  true,
	"GET":  true,
}

type Command struct {
	Data         []string
	CommandName  string
	Arguments    []interface{}
	ExpectedSize int
	ExpectedType string
	CurrentSize  int
	Completed    bool
}

func NewParser(buffer *bufio.Reader) *Parser {
	return &Parser{
		Source:  buffer,
		Start:   0,
		Current: 0,
		Tokens:  make([]Token, 0),
	}
}

func (p *Parser) Parse() error {

	var currentCommand Command = Command{}
	for {
		line, err := p.Source.ReadString('\n')
		if err != nil {
			return err
		}
		dataType := line[0]

		switch dataType {
		case '*':
			size, err := strconv.Atoi(string(line[1]))
			if err != nil {
				return err
			}
			currentCommand = Command{
				Data:         make([]string, 0, size),
				ExpectedSize: size,
				CurrentSize:  0,
				Arguments:    make([]interface{}, 0, size-1),
				Completed:    false,
			}
		case '$':
			if currentCommand.ExpectedSize == 0 {
				return fmt.Errorf("-Command in bad format")
			}
			if currentCommand.CurrentSize == 0 {
				currentCommand.ExpectedType = "command"
			} else {
				currentCommand.ExpectedType = "string"
			}
			currentCommand.CurrentSize++
		default:
			if currentCommand.ExpectedSize == 0 {
				return fmt.Errorf("-Command in bad format")
			}

			if currentCommand.ExpectedType == "command" {
				uppercaseLine := strings.ToUpper(strings.Trim(line, "\r\n"))
				if _, ok := SupportedCommands[uppercaseLine]; !ok {
					return fmt.Errorf("-Unsupported Command")
				}
				currentCommand.CommandName = uppercaseLine
				if currentCommand.CurrentSize == currentCommand.ExpectedSize {
					currentCommand.Completed = true
					p.CurrentCommand = &currentCommand
					return nil
				}
			}

			if currentCommand.ExpectedType == "string" {
				currentCommand.Arguments = append(currentCommand.Arguments, strings.Trim(line, "\r\n"))
				if currentCommand.CurrentSize == currentCommand.ExpectedSize {
					currentCommand.Completed = true
					p.CurrentCommand = &currentCommand
					return nil
				}

			}
		}
	}

}
