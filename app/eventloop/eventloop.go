package eventloop

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/commands"
)

type EventLoop struct {
	Tasks chan commands.Command
	stop  chan bool
}

func NewEventLoop() *EventLoop {
	return &EventLoop{
		Tasks: make(chan commands.Command),
	}
}

func (e *EventLoop) Run() {

	for {
		select {
		case task := <-e.Tasks:
			output, err := task.ExecuteCommand()
			resultChan := task.GetResponseChan()
			if err != nil {
				serializedError := commands.SerializeOutput(err, true)
				resultChan <- serializedError
				continue
			}
			outputSerialized := commands.SerializeOutput(output, false)
			if outputSerialized == nil {
				serializedError := commands.SerializeOutput(fmt.Errorf("unsupported protocol type"), true)
				resultChan <- serializedError
				continue
			}
			resultChan <- outputSerialized

		case stop := <-e.stop:
			if stop {
				return
			}
		}
	}

}
