package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/db"
)

type baseCommand struct {
	db         *db.Db
	args       []string
	Response   chan []byte
	isBlocking bool
	callback   Command
}

func (c *baseCommand) GetResponseChan() chan []byte {
	return c.Response
}
func (c *baseCommand) SetResponseChan(newChan chan []byte) {
	c.Response = newChan
}
func (c *baseCommand) IsBlocking() bool {
	return c.isBlocking
}
func (c *baseCommand) Callback() Command {
	return c.callback
}

type Command interface {
	ExecuteCommand() (any, error)
	IsBlocking() bool
	GetResponseChan() chan []byte
	Callback() Command
	SetResponseChan(newChan chan []byte)
}

func NewCommand(name string, db *db.Db, args []string) (Command, error) {
	b := baseCommand{
		db:       db,
		args:     args,
		Response: make(chan []byte),
	}
	switch strings.ToUpper(name) {
	case "PING":
		// PING doesn't need the db, so we can overwrite the base
		return &PingCommand{baseCommand: baseCommand{args: args, Response: b.Response}}, nil
	case "ECHO":
		return &EchoCommand{baseCommand: baseCommand{args: args, Response: b.Response}}, nil
	case "GET":
		return &GetCommand{baseCommand: b}, nil
	case "SET":
		return &SetCommand{baseCommand: b}, nil
	case "RPUSH":
		return &RPUSHCommand{baseCommand: b}, nil
	case "LPUSH":
		return &LPUSHCommand{baseCommand: b}, nil
	case "LLEN":
		return &LLENCommand{baseCommand: b}, nil
	case "LPOP":
		return &LPOPCommand{baseCommand: b}, nil
	case "BLPOP":
		b.isBlocking = true
		return &BLPOPCommand{baseCommand: b}, nil
	case "LRANGE":
		return &LRANGECommand{baseCommand: b}, nil
	default:
		return nil, fmt.Errorf("unknown command '%s'", name)
	}
}

type PingCommand struct {
	baseCommand
}

func (c *PingCommand) ExecuteCommand() (any, error) {
	args := c.args
	if len(args) != 1 {
		return "", fmt.Errorf("wrong number of arguments for 'PING' command")
	}
	return "PONG", nil
}

type EchoCommand struct {
	baseCommand
}

func (c *EchoCommand) ExecuteCommand() (any, error) {
	args := c.args
	if len(args) != 2 {
		return "", fmt.Errorf("wrong number of arguments for 'ECHO' command")
	}
	return args[1], nil
}

type GetCommand struct {
	baseCommand
}

func (c *GetCommand) ExecuteCommand() (any, error) {
	args := c.args
	if len(args) != 2 {
		return "", fmt.Errorf("wrong number of arguments for 'GET' command")
	}
	key := args[1]

	val, ok := c.db.GetValue(key)
	if !ok {
		return nil, nil
	}
	return val, nil
}

type SetCommand struct {
	baseCommand
}

func (c *SetCommand) ExecuteCommand() (any, error) {
	args := c.args
	if len(args) < 3 || len(args) != 5 {
		return "", fmt.Errorf("wrong number of arguments for 'SET' command")
	}
	key := args[1]
	value := args[2]

	dbVal := db.MapValue{Value: value}

	if len(args) == 5 && strings.ToUpper(args[3]) == "PX" {
		milliseconds, err := strconv.ParseInt(args[4], 10, 64)
		if err != nil {
			return "", fmt.Errorf("value is not an integer or out of range")
		}
		dbVal.HasExpiryDate = true
		dbVal.ExpireAt = time.Now().Add(time.Duration(milliseconds) * time.Millisecond)
	}

	c.db.DbMap[key] = &dbVal

	return "OK", nil
}

type RPUSHCommand struct {
	baseCommand
}

func (c *RPUSHCommand) ExecuteCommand() (any, error) {
	args := c.args
	if len(args) < 3 {
		return "", fmt.Errorf("wrong number of arguments for 'RPUSH' command")
	}
	key := args[1]

	// TO DO - refactor this a bit to use the GetValue method
	_, ok := c.db.DbMap[key]

	if !ok {
		c.db.DbMap[key] = &db.MapValue{
			Value: make([]string, 0),
			SetAt: time.Now(),
		}
		if _, ok := c.db.ListChannels[key]; !ok {
			c.db.ListChannels[key] = make(chan bool, 1)
		}
	}
	val := c.db.DbMap[key]
	for i := 2; i < len(args); i++ {
		val.Value = append(val.Value.([]string), args[i])
	}

	listSize := len(val.Value.([]string))
	if val.HasExpiryDate && time.Now().After(val.ExpireAt) {
		delete(c.db.DbMap, key)
		return "-1", nil
	}
	select {
	case c.db.ListChannels[key] <- true:
	default:
	}

	return listSize, nil
}

type LPUSHCommand struct {
	baseCommand
}

func (c *LPUSHCommand) ExecuteCommand() (any, error) {
	args := c.args
	if len(args) < 3 {
		return "", fmt.Errorf("wrong number of arguments for 'LPUSH' command")
	}
	key := args[1]

	// TO DO - refactor this a bit to use the GetValue method
	_, ok := c.db.DbMap[key]

	if !ok {
		c.db.DbMap[key] = &db.MapValue{
			Value: make([]string, 0),
			SetAt: time.Now(),
		}
		if _, ok := c.db.ListChannels[key]; !ok {
			c.db.ListChannels[key] = make(chan bool, 1)
		}
	}
	val := c.db.DbMap[key]
	for i := 2; i < len(args); i++ {
		val.Value = append([]string{args[i]}, val.Value.([]string)...)
	}

	listSize := len(val.Value.([]string))
	if val.HasExpiryDate && time.Now().After(val.ExpireAt) {
		delete(c.db.DbMap, key)
		return "-1", nil
	}
	select {
	case c.db.ListChannels[key] <- true:
	default:
	}
	return listSize, nil
}

type LLENCommand struct {
	baseCommand
}

func (c *LLENCommand) ExecuteCommand() (any, error) {
	args := c.args
	if len(args) != 2 {
		return "", fmt.Errorf("wrong number of arguments for 'LLEN' command")
	}
	key := args[1]

	// TO DO - refactor this a bit to use the GetValue method
	val, ok := c.db.DbMap[key]

	if !ok {
		return 0, nil
	}
	valAsList, ok := val.Value.([]string)
	if !ok {
		return "", fmt.Errorf("value not a list")
	}

	listSize := len(valAsList)
	if val.HasExpiryDate && time.Now().After(val.ExpireAt) {
		delete(c.db.DbMap, key)
		return "-1", nil
	}

	return listSize, nil
}

type LPOPCommand struct {
	baseCommand
}

func (c *LPOPCommand) ExecuteCommand() (any, error) {
	fmt.Println("executing  LPOP")
	args := c.args
	if len(args) > 3 {
		return "", fmt.Errorf("wrong number of arguments for 'LLEN' command")
	}
	key := args[1]
	var numberOfElements = 1
	var err error
	if len(args) == 3 {
		numberOfElements, err = strconv.Atoi(args[2])
		if err != nil {
			return "", fmt.Errorf("argument to pop must be an integer")
		}
	}

	// TO DO - refactor this a bit to use the GetValue method
	val, ok := c.db.DbMap[key]

	if !ok {
		return 0, nil
	}
	valAsList, ok := val.Value.([]string)
	if !ok {
		return "", fmt.Errorf("value not a list")
	}
	var first any
	if numberOfElements == 1 {
		first = valAsList[0]
	} else {
		first = valAsList[:numberOfElements]
	}
	val.Value = valAsList[numberOfElements:]

	if val.HasExpiryDate && time.Now().After(val.ExpireAt) {
		delete(c.db.DbMap, key)
		delete(c.db.ListChannels, key)
		return "-1", nil
	}
	if len(valAsList)-numberOfElements == 0 {
		delete(c.db.DbMap, key)
		delete(c.db.ListChannels, key)
	}

	return first, nil
}

type BLPOPCommand struct {
	baseCommand
}

func (c *BLPOPCommand) ExecuteCommand() (any, error) {
	fmt.Println("executing  BLPOP")
	args := c.args
	if len(args) != 3 {
		return "", fmt.Errorf("wrong number of arguments for 'BLPOP' command")
	}
	key := args[1]
	var timeout = 0.0
	var err error
	timeout, err = strconv.ParseFloat(args[2], 32)
	if err != nil {
		return "", fmt.Errorf("argument to blpop must be an integer")
	}

	// TO DO - refactor this a bit to use the GetValue method
	if _, ok := c.db.ListChannels[key]; !ok {
		c.db.ListChannels[key] = make(chan bool, 1)
	}
	if timeout == 0 {
		<-c.db.ListChannels[key]
	} else {
		blocking := true
		startTime := time.Now()
		for blocking {
			select {
			case <-c.db.ListChannels[key]:
				blocking = false
			default:
				if time.Since(startTime) > time.Duration(timeout*float64(time.Second)) {
					blocking = false
					return []string{"-1"}, nil
				}
			}
		}
	}
	c.callback, _ = NewCallback("BLPOP", c.db, []string{"LPOP", key})
	c.callback.SetResponseChan(c.Response)

	return nil, nil
}

type LRANGECommand struct {
	baseCommand
}

func (c *LRANGECommand) ExecuteCommand() (any, error) {
	args := c.args
	if len(args) != 4 {
		return "", fmt.Errorf("wrong number of arguments for 'LRANGE' command")
	}
	key := args[1]

	val, ok := c.db.GetValue(key)
	if !ok {
		return []string{}, nil
	}

	startIndex, err := strconv.Atoi(args[2])
	if err != nil {
		return "", fmt.Errorf("wrong value for argument,expected integer")
	}
	stopIndex, err := strconv.Atoi(args[3])
	if err != nil {
		return "", fmt.Errorf("wrong value for argument,expected integer")
	}

	valAsList, ok := val.([]string)
	if !ok {
		return "", fmt.Errorf("value not a list")
	}
	if startIndex < 0 {
		startIndex = len(valAsList) + startIndex
	}
	if stopIndex < 0 {
		stopIndex = len(valAsList) + stopIndex
	}

	if startIndex < 0 {
		startIndex = 0
	}
	if stopIndex < 0 {
		stopIndex = 0
	}
	if startIndex >= len(valAsList) {
		return []string{}, nil
	}
	if stopIndex >= len(valAsList) {
		stopIndex = len(valAsList) - 1
	}
	if startIndex > stopIndex {
		return []string{}, nil
	}
	return valAsList[startIndex : stopIndex+1], nil
}
