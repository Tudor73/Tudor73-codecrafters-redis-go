package commands

type MULTICommand struct {
	baseCommand
}

func (c *MULTICommand) ExecuteCommand() (string, error) {
	return OK, nil
}
