package command

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/db"
)

type Command interface {
	ExecuteCommand(args []string) (any, error)
}

func NewCommand(name string, db *db.Db) (Command, error) {
	switch strings.ToUpper(name) {
	case "PING":
		return &PingCommand{}, nil
	case "ECHO":
		return &EchoCommand{}, nil
	case "GET":
		return &GetCommand{db: db}, nil
	case "SET":
		return &SetCommand{db: db}, nil
	case "RPUSH":
		return &RPUSHCommand{db: db}, nil
	case "LRANGE":
		return &LRANGECommand{db: db}, nil
	default:
		return nil, fmt.Errorf("unknown command '%s'", name)
	}
}

type PingCommand struct{}

func (c *PingCommand) ExecuteCommand(args []string) (any, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("wrong number of arguments for 'PING' command")
	}
	return "PONG", nil
}

type EchoCommand struct{}

func (c *EchoCommand) ExecuteCommand(args []string) (any, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("wrong number of arguments for 'ECHO' command")
	}
	return args[1], nil
}

type GetCommand struct {
	db *db.Db
}

func (c *GetCommand) ExecuteCommand(args []string) (any, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("wrong number of arguments for 'GET' command")
	}
	key := args[1]

	c.db.Mu.Lock()
	val, ok := c.db.GetValue(key)
	c.db.Mu.Unlock()
	if !ok {
		return nil, nil
	}
	return val, nil
}

type SetCommand struct {
	db *db.Db
}

func (c *SetCommand) ExecuteCommand(args []string) (any, error) {
	if len(args) < 3 {
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

	c.db.Mu.Lock()
	c.db.DbMap[key] = &dbVal
	c.db.Mu.Unlock()

	return "OK", nil
}

type RPUSHCommand struct {
	db *db.Db
}

func (c *RPUSHCommand) ExecuteCommand(args []string) (any, error) {
	if len(args) < 3 {
		return "", fmt.Errorf("wrong number of arguments for 'RPUSH' command")
	}
	key := args[1]

	// TO DO - refactor this a bit to use the GetValue method
	c.db.Mu.Lock()
	_, ok := c.db.DbMap[key]

	if !ok {
		c.db.DbMap[key] = &db.MapValue{
			Value: make([]string, 0),
			SetAt: time.Now(),
		}
	}
	val := c.db.DbMap[key]
	for i := 2; i < len(args); i++ {
		val.Value = append(val.Value.([]string), args[i])
	}

	listSize := len(val.Value.([]string))
	c.db.Mu.Unlock()
	if val.HasExpiryDate && time.Now().After(val.ExpireAt) {
		c.db.Mu.Lock()
		delete(c.db.DbMap, key)
		c.db.Mu.Unlock()
		return "-1", nil
	}

	return listSize, nil
}

type LRANGECommand struct {
	db *db.Db
}

func (c *LRANGECommand) ExecuteCommand(args []string) (any, error) {
	if len(args) != 4 {
		return "", fmt.Errorf("wrong number of arguments for 'LRANGE' command")
	}
	key := args[1]

	c.db.Mu.Lock()
	val, ok := c.db.GetValue(key)
	c.db.Mu.Unlock()
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
