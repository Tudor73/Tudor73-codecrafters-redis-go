package eventloop

import (
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
						serializedError := commands.EncodeError(err.Error())
						resultChan <- []byte(serializedError)
						return
					}
					if output == "" && task.Callback() != nil {
						e.Callbacks <- task.Callback()
						return
					}
					if output == "" {
						serializedError := commands.EncodeError("Unsupported protocol")
						resultChan <- []byte(serializedError)
						return
					}
					resultChan <- []byte(output)

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
		serializedError := commands.EncodeError(err.Error())
		resultChan <- []byte(serializedError)
		return
	}
	if output == "" {
		serializedError := commands.EncodeError("Unsupported protocol")
		resultChan <- []byte(serializedError)
		return
	}
	resultChan <- []byte(output)
}
