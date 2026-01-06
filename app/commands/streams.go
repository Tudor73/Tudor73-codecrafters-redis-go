package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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
	var err error
	if !ok {
		id, err = validateId(id, nil)
		if err != nil {
			return "", err
		}
		s.db.DbMap[key] = &db.MapValue{
			Value: []StreamValue{StreamValue{
				Id:     id,
				Fields: make(map[string]any),
			}},
		}
	} else {
		id, err = validateId(id, &val.Value.([]StreamValue)[len(val.Value.([]StreamValue))-1])
		if err != nil {
			return "", err
		}
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

func validateId(id string, lastEntry *StreamValue) (string, error) {

	miliseconds, sequenceNumber, err := parseId(id)
	if err != nil {
		return "", err
	}
	prevMiliseconds, prevNumber := 0, 0
	if lastEntry != nil {
		prevMiliseconds, prevNumber, _ = parseId(lastEntry.Id)
	}
	if sequenceNumber == -1 {
		if miliseconds == prevMiliseconds {
			sequenceNumber = prevNumber + 1
		} else {
			sequenceNumber = 0
		}
	}
	if miliseconds == 0 && sequenceNumber == 0 {
		return "", fmt.Errorf("ERR The ID specified in XADD must be greater than 0-0")
	}
	if miliseconds < prevMiliseconds {
		return "", fmt.Errorf("ERR The ID specified in XADD is equal or smaller than the target stream top item")
	}
	if miliseconds == prevMiliseconds && sequenceNumber == prevNumber {
		return "", fmt.Errorf("ERR The ID specified in XADD is equal or smaller than the target stream top item")
	}
	return fmt.Sprintf("%d-%d", miliseconds, sequenceNumber), nil
}

func parseId(id string) (int, int, error) {
	values := strings.Split(id, "-")
	if len(values) > 2 {
		return 0, 0, fmt.Errorf("ERR the id specified is not in the right format")
	}
	if len(values) == 1 {
		if values[0] != "*" {
			return 0, 0, fmt.Errorf("ERR the id specified is not in the right format")
		}
		values[0] = strconv.FormatInt(time.Now().UnixMilli(), 10)
		values = append(values, "*")
	}

	if values[1] == "*" {
		values[1] = "-1"
	}

	miliseconds, err := strconv.Atoi(values[0])
	if err != nil {
		return 0, 0, fmt.Errorf("ERR invalid format")
	}
	sequenceNumber, err := strconv.Atoi(values[1])
	if err != nil {
		return 0, 0, fmt.Errorf("ERR invalid format")
	}

	return miliseconds, sequenceNumber, nil
}
