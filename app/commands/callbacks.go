package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/db"
)

func NewCallback(name string, db *db.Db, args []string) (Command, error) {
	b := baseCommand{
		db:       db,
		args:     args,
		Response: make(chan []byte),
	}
	switch strings.ToUpper(name) {
	case "BLPOP":
		return &BLPOPCallback{baseCommand: b}, nil
	default:
		return nil, fmt.Errorf("unknown command '%s'", name)
	}
}

type BLPOPCallback struct {
	baseCommand
}

func (c *BLPOPCallback) ExecuteCommand() (any, error) {
	args := c.args
	key := args[1]

	val, ok := c.db.DbMap[key]

	if !ok {
		return []string{}, nil
	}
	valAsList, ok := val.Value.([]string)
	if !ok {
		return "", fmt.Errorf("wrong number of arguments for 'LLEN' command")
	}
	first := valAsList[0]
	val.Value = valAsList[1:]

	if val.HasExpiryDate && time.Now().After(val.ExpireAt) {
		delete(c.db.DbMap, key)
		delete(c.db.ListChannels, key)
		return "-1", nil
	}
	if len(valAsList)-1 == 0 {
		delete(c.db.DbMap, key)
		delete(c.db.ListChannels, key)
	} else {
		c.db.ListChannels[key] <- true
	}

	result := []string{key, first}
	return result, nil
}
