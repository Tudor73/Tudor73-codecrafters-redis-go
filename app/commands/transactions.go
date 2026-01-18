package commands

import "fmt"

type MULTICommand struct {
	baseCommand
}

func (c *MULTICommand) ExecuteCommand() (string, error) {
	return OK, nil
}

type EXECCommand struct {
	baseCommand
}

func (c *EXECCommand) ExecuteCommand() (string, error) {
	return "", fmt.Errorf("ERR EXEC without MULTI")
}
