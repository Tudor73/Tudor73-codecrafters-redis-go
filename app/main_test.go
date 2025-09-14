package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var client *redis.Client
var redisContainer testcontainers.Container

func TestMain(m *testing.M) {
	ctx := context.Background()
	wd, _ := os.Getwd()
	parent := filepath.Dir(wd)
	fmt.Println(parent)
	var err error
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    parent,
			Dockerfile: "Dockerfile",
			Repo:       "tudor73",
			Tag:        "redis2",
		},
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Logs from your program will appear here!"),
	}
	redisContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	endpoint, err := redisContainer.Endpoint(ctx, "")
	if err != nil {
		log.Fatal(err.Error())
	}

	client = redis.NewClient(&redis.Options{
		Addr: endpoint,
	})
	exitCode := m.Run()
	if err := redisContainer.Terminate(ctx); err != nil {
		log.Fatalf("Could not stop redis container: %s", err)
	}
	os.Exit(exitCode)
}

// TestSetCommand uses the shared client to test the SET command.
func TestSetCommand(t *testing.T) {
	ctx := context.Background()
	key := "mykey1"
	value := "myvalue"

	err := client.Set(ctx, key, value, 0).Err()
	require.NoError(t, err, "Failed to set key")

	actualValue, err := client.Get(ctx, key).Result()
	require.NoError(t, err)
	require.Equal(t, value, actualValue)
}

func TestBLPOPCommand(t *testing.T) {
	ctx := context.Background()
	state, err := redisContainer.State(ctx)
	require.NoError(t, err)
	t.Logf("TestGetCommand: Container is running: %t", state.Running)
	key := "mykey"

	var resultChan = make(chan *redis.StringSliceCmd)
	go func() {
		resultChan <- client.BLPop(ctx, 2*time.Second, key)
	}()

	endpoint, err := redisContainer.Endpoint(ctx, "")
	if err != nil {
		log.Fatal(err.Error())
	}
	client_temp := redis.NewClient(&redis.Options{
		Addr: endpoint,
	})
	client_temp.RPush(ctx, key, "strawberry")

	result := <-resultChan
	val, err := result.Result()
	assert.NoError(t, err)
	assert.Equal(t, []string{key, "strawberry"}, val)
}
