package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/db"
)

type StreamValue struct {
	Id     string
	Fields map[string]any
}

type StreamAddCommand struct {
	baseCommand
}

func (s *StreamAddCommand) ExecuteCommand() (string, error) {
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
			Value: []StreamValue{{
				Id:     id,
				Fields: mapValues,
			}},
		}
	} else {
		id, err = validateId(id, &val.Value.([]StreamValue)[len(val.Value.([]StreamValue))-1])
		if err != nil {
			return "", err
		}
		val.Value = append(val.Value.([]StreamValue), StreamValue{Id: id, Fields: mapValues})
	}
	return BulkString(id), nil

}

type StreamRangeCommand struct {
	baseCommand
}

func (s *StreamRangeCommand) ExecuteCommand() (string, error) {
	args := s.args

	if len(args) != 4 {
		return "", fmt.Errorf("wrong number of arguments for XRANGE command")
	}
	key := args[1]
	from, to := args[2], args[3]
	val, ok := s.db.DbMap[key]
	if !ok {
		return "", fmt.Errorf("key not found")
	}

	valAsList, ok := val.Value.([]StreamValue)
	if !ok {
		return "", fmt.Errorf("key not a stream")
	}
	var result []any
	var fromIndex int
	for i, v := range valAsList {
		if v.Id == from {
			result = append(result, v)
			fromIndex = i
			break
		}
	}
	if len(result) == 0 {
		return "", fmt.Errorf("id not found")
	}
	for j := fromIndex + 1; j < len(valAsList); j++ {
		result = append(result, valAsList[j])
		if valAsList[j].Id == to {
			break
		}
	}

	encodedArray, err := Encode(result)
	if err != nil {
		return "", err
	}
	return encodedArray, nil
}

func parseMapValues(args []string) map[string]any {
	mapValues := make(map[string]any)
	for i := 3; i < len(args); i += 2 {
		mapValues[args[i]] = args[i+1]
	}
	return mapValues
}

type StreamId struct {
	Ms  uint64
	Seq int
}

func (s StreamId) Compare(other StreamId) int {
	if s.Ms > other.Ms {
		return 1
	}
	if s.Ms < other.Ms {
		return -1
	}
	if s.Seq > other.Seq {
		return 1
	}
	if s.Seq < other.Seq {
		return -1
	}
	return 0
}

func validateId(id string, lastEntry *StreamValue) (string, error) {

	miliseconds, sequenceNumber, autoSeq, err := parseId(id)
	if err != nil {
		return "", err
	}
	var prev StreamId
	if lastEntry != nil {
		// Ideally lastEntry already has a StreamID struct to avoid re-parsing
		pMs, pSeq, _, _ := parseId(lastEntry.Id)
		prev = StreamId{Ms: uint64(pMs), Seq: pSeq}
	}

	if autoSeq {
		if miliseconds == prev.Ms {
			sequenceNumber = prev.Seq + 1
		} else {
			sequenceNumber = 0
		}
	}

	current := StreamId{Ms: miliseconds, Seq: sequenceNumber}

	if current.Ms == 0 && current.Seq == 0 {
		return "", fmt.Errorf("ERR The ID specified in XADD must be greater than 0-0")
	}
	if lastEntry != nil && current.Compare(prev) <= 0 {
		return "", fmt.Errorf("ERR The ID specified in XADD is equal or smaller than the target stream top item")
	}

	return fmt.Sprintf("%d-%d", miliseconds, sequenceNumber), nil
}

func parseId(id string) (uint64, int, bool, error) {
	values := strings.Split(id, "-")
	if len(values) > 2 {
		return 0, 0, false, fmt.Errorf("ERR the id specified is not in the right format")
	}
	if len(values) == 1 {
		if values[0] != "*" {
			return 0, 0, false, fmt.Errorf("ERR the id specified is not in the right format")
		}
		values[0] = strconv.FormatInt(time.Now().UnixMilli(), 10)
		values = append(values, "*")
	}

	miliseconds, err := strconv.Atoi(values[0])
	if err != nil {
		return 0, 0, false, fmt.Errorf("ERR invalid format")
	}
	if values[1] == "*" {
		return uint64(miliseconds), 0, true, nil
	}
	sequenceNumber, err := strconv.Atoi(values[1])
	if err != nil {
		return 0, 0, false, fmt.Errorf("ERR invalid format")
	}

	return uint64(miliseconds), sequenceNumber, false, nil
}
