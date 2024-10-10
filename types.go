package queen

import (
	"context"
	"crypto"
	"github.com/hibiken/asynq"
)

type CubeType struct {
	Value string
}

type CubeConfig struct {
	Id        string
	CubeType  CubeType
	CubeName  string
	PublicKey crypto.PublicKey
}

type Task struct {
	taskType    string
	taskHandler func(ctx context.Context, t *asynq.Task) error
}

type Config struct {
	CubeConfig
	RedisAddress  string
	RedisPassword string
	RedisDB       int
	Tasks         map[string]Task
}

type Message struct {
	MessageType string
	Message     interface{}
}
