package parser

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
)

type Parser struct {
	Source  *bufio.Reader
	Start   int
	Current int
	Err     error
}

func NewParser(buffer *bufio.Reader) *Parser {
	return &Parser{
		Source:  buffer,
		Start:   0,
		Current: 0,
	}
}

var (
	ErrInvalidSyntax      = errors.New("invalid syntax")
	ErrUnsupportedType    = errors.New("unsupported RESP type")
	ErrInvalidBulkStrSize = errors.New("invalid bulk string size")
	ErrInvalidArraySize   = errors.New("invalid array size")
)

func (p *Parser) Parse() (any, error) {

	dataType, err := p.Source.ReadByte()
	if err != nil {
		return nil, err
	}

	switch dataType {
	case '*':
		return p.ParseArray()
	case '$':
		return p.ParseBulkString()
	default:
		return nil, ErrInvalidSyntax
	}
}

func (p *Parser) ParseInteger() (int, error) {
	line, err := p.readLine()
	if err != nil {
		return 0, err
	}
	number, err := strconv.Atoi(line)
	if err != nil {
		return 0, err
	}
	return number, nil
}

func (p *Parser) readLine() (string, error) {
	line, err := p.Source.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.Trim(line, "\r\n"), nil
}

func (p *Parser) ParseBulkString() (any, error) {
	length, err := p.ParseInteger()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, length)
	_, err = io.ReadFull(p.Source, buf)
	if err != nil {
		return nil, ErrInvalidBulkStrSize
	}

	// this reads the '\r\n' before going to the next byte is read
	if _, err := p.readLine(); err != nil {
		return nil, err
	}
	return string(buf), nil

}

func (p *Parser) ParseArray() (any, error) {
	length, err := p.ParseInteger()
	if err != nil {
		return nil, err
	}

	array := make([]any, length)
	for i := 0; i < length; i++ {
		value, err := p.Parse()
		if err != nil {
			return nil, err
		}
		array[i] = value
	}
	return array, nil
}
