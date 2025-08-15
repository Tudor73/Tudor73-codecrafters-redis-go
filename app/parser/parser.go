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
}

type Command struct {
	Data         []string
	CommandName  string
	Argument     interface{}
	ExpectedSize int
	ExpectedType string
	CurrentSize  int
	Completed    bool
}

type ProtocolTypeHandler func(*Parser) error

var typeIndentifiers = map[string]ProtocolTypeHandler{
	"+": HandleString,
	"-": HandleError,
	":": HandleInteger,
	"$": HandleBulkStrings,
	"*": HandleArrays,
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
			}

			if currentCommand.ExpectedType == "string" {
				currentCommand.Argument = strings.Trim(line, "\r\n")
				if currentCommand.CurrentSize == currentCommand.ExpectedSize {
					currentCommand.Completed = true
					p.CurrentCommand = &currentCommand
					return nil
				}

			}
		}
	}

}

func HandleString(p *Parser) error {
	return nil
}

func HandleError(p *Parser) error {

	return nil
}

func HandleInteger(p *Parser) error {

	return nil
}
func HandleBulkStrings(p *Parser) error {
	return nil
}

func HandleArrays(p *Parser) error {
	return nil
}

func (p *Parser) isAtEnd() bool {
	return false
}

func (p *Parser) peek() string {

	if p.isAtEnd() {
		return ""
	}
	return ""
}

func (p *Parser) advance() string {

	return ""
}

func (p *Parser) AddToken(dataType string, value interface{}) {
	p.Tokens = append(p.Tokens, Token{DataType: dataType, Value: value})
}
