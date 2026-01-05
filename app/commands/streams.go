package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/db"
)

type StreamCommand struct {
	baseCommand
}

type StreamValue struct {
	Id     string
	Fields map[string]any
}

func (s *StreamCommand) ExecuteCommand() (any, error) {
	args := s.args

	if len(args)%2 != 1 || len(args) < 3 {
		return "", fmt.Errorf("wrong number of arguments for XADD command")
	}
	key := args[1]
	id := args[2]

	mapValues := parseMapValues(args)

	val, ok := s.db.DbMap[key]
	if !ok {
		s.db.DbMap[key] = &db.MapValue{
			Value: []StreamValue{StreamValue{
				Id:     id,
				Fields: make(map[string]any),
			}},
		}
	} else {
		val.Value = append(val.Value.([]StreamValue), StreamValue{Id: id, Fields: mapValues})
	}
	return id, nil

}

func parseMapValues(args []string) map[string]any {
	mapValues := make(map[string]any)
	for i := 3; i < len(args); i += 2 {
		mapValues[args[i]] = args[i+1]
	}
	return mapValues
}
