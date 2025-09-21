package eventloop

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/commands"
)

type EventLoop struct {
	Tasks     chan commands.Command
	Callbacks chan commands.Command
	stop      chan bool
}

func NewEventLoop() *EventLoop {
	return &EventLoop{
		Tasks:     make(chan commands.Command),
		Callbacks: make(chan commands.Command),
	}
}

func (e *EventLoop) Run() {
	for {
		select {
		case task := <-e.Tasks:
			if task.IsBlocking() {
				go func() {
					output, err := task.ExecuteCommand()
					resultChan := task.GetResponseChan()
					if err != nil {
						serializedError := commands.SerializeOutput(task.GetName(), err, true)
						resultChan <- serializedError
						return
					}
					if output == nil && task.Callback() != nil {
						e.Callbacks <- task.Callback()
						return
					}
					outputSerialized := commands.SerializeOutput(task.GetName(), output, false)
					if outputSerialized == nil {
						serializedError := commands.SerializeOutput("", fmt.Errorf("unsupported protocol type"), true)
						resultChan <- serializedError
						return
					}
					resultChan <- outputSerialized

				}()
			} else {
				handleTask(task)
			}

		case task := <-e.Callbacks:
			handleTask(task)
		case stop := <-e.stop:
			if stop {
				return
			}
		}
	}

}

func handleTask(task commands.Command) {
	output, err := task.ExecuteCommand()
	resultChan := task.GetResponseChan()
	if err != nil {
		serializedError := commands.SerializeOutput(task.GetName(), err, true)
		resultChan <- serializedError
		return
	}
	outputSerialized := commands.SerializeOutput(task.GetName(), output, false)
	if outputSerialized == nil {
		serializedError := commands.SerializeOutput("", fmt.Errorf("unsupported protocol type"), true)
		resultChan <- serializedError
		return
	}
	resultChan <- outputSerialized
}
