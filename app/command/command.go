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
	val, ok := c.db.DbMap[key]
	c.db.Mu.Unlock()

	if !ok {
		return "-1", nil
	}

	if val.HasExpiryDate && time.Now().After(val.ExpireAt) {
		c.db.Mu.Lock()
		delete(c.db.DbMap, key)
		c.db.Mu.Unlock()
		return nil, nil
	}

	return val.Value, nil
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
