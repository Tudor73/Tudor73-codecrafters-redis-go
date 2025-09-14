package commands

import (
	"errors"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/db"
	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	db := db.NewDb()
	testCases := []struct {
		args           []string
		expectedOutput any
		expectedError  error
		success        bool
	}{
		{
			args:           []string{"SET", "foo", "strawberry"},
			expectedOutput: "OK",
			success:        true,
		},
	}

	for _, tt := range testCases {
		command, _ := NewCommand("SET", db, tt.args)
		output, err := command.ExecuteCommand()
		assert.Equal(t, tt.expectedError, err)
		assert.Equal(t, tt.expectedOutput, output)
		_, ok := db.GetValue("foo")
		assert.Equal(t, tt.success, ok)
		db.DelValue("foo")
	}
}

func TestSetWithTimeout(t *testing.T) {
	db := db.NewDb()
	testCases := []struct {
		args           []string
		expectedOutput any
		expectedError  error
		success        bool
	}{
		{
			args:           []string{"SET", "foo", "strawberry", "PX", "1"},
			expectedOutput: "OK",
			success:        true,
		},
		{
			args:           []string{"SET", "foo", "strawberry", "PX"},
			expectedOutput: "",
			expectedError:  errors.New("wrong number of arguments for 'SET' command"),
			success:        false,
		},
	}

	for _, tt := range testCases {
		command, _ := NewCommand("SET", db, tt.args)
		output, err := command.ExecuteCommand()
		assert.Equal(t, tt.expectedError, err)
		assert.Equal(t, tt.expectedOutput, output)
		_, ok := db.GetValue("foo")
		assert.Equal(t, tt.success, ok)
		if tt.success == true {
			time.Sleep(time.Second)
			_, ok = db.GetValue("foo")
			assert.Equal(t, false, ok)
		}
		db.DelValue("foo")
	}
}
func TestRPUSHCommand(t *testing.T) {
	db := db.NewDb()

	testCases := []struct {
		args           []string
		expectedOutput any
	}{
		{
			args:           []string{"RPUSH", "foo", "strawberry", "apple"},
			expectedOutput: 2,
		},
		{
			args:           []string{"RPUSH", "foo", "orange"},
			expectedOutput: 3,
		},
	}

	for _, tt := range testCases {
		command, _ := NewCommand("RPUSH", db, tt.args)
		output, err := command.ExecuteCommand()
		assert.NoError(t, err)
		assert.Equal(t, tt.expectedOutput, output)
	}
}

func TestLPOPCommand(t *testing.T) {
	db := db.NewDb()
	db.SetValue("foo", []string{"strawberry", "apple", "orange"})
	testCases := []struct {
		args           []string
		expectedOutput any
	}{
		{
			args:           []string{"LPOP", "foo", "2"},
			expectedOutput: []string{"strawberry", "apple"},
		},
		{
			args:           []string{"LPOP", "foo", "1"},
			expectedOutput: "orange",
		},
	}

	for _, tt := range testCases {
		command, err := NewCommand("LPOP", db, tt.args)
		assert.NoError(t, err)
		output, err := command.ExecuteCommand()
		assert.NoError(t, err)
		assert.Equal(t, tt.expectedOutput, output)
	}
	_, ok := db.GetValue("foo")
	assert.Equal(t, false, ok)
}
