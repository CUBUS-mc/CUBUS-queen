package queen

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/hibiken/asynq"
	"os"
)

func Test() string {
	return "This is a test"
}

type Queen struct {
	config         Config
	privateKey     crypto.PrivateKey
	messageChannel chan Message
	asynqClient    *asynq.Client
}

func (q *Queen) NewQueen(config Config) *Queen {
	q.config = config
	if q.config.PublicKey == nil {
		q.generateKeyPair()
	} else {
		q.loadPrivateKey()
	}
	q.messageChannel = make(chan Message)
	go q.startServer()
	go q.startClient()
	return q
}

func (q *Queen) generateKeyPair() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	q.privateKey = privateKey
	q.config.PublicKey = &privateKey.PublicKey

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	err = os.WriteFile(q.config.Id+".pem", privateKeyPEM, 0600)
	if err != nil {
		panic(err)
	}
	q.messageChannel <- Message{
		MessageType: "UPDATE PUBLIC KEY",
		Message:     q.config.PublicKey,
	}
}

func (q *Queen) loadPrivateKey() {
	privateKeyPEM, err := os.ReadFile(q.config.Id + ".pem")
	if err != nil {
		panic(err)
	}

	block, _ := pem.Decode(privateKeyPEM)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		panic(errors.New("failed to decode PEM block containing private key"))
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	q.privateKey = privateKey
}

func (q *Queen) startServer() {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     q.config.RedisAddress,
			Password: q.config.RedisPassword,
			DB:       q.config.RedisDB,
		},
		asynq.Config{
			Concurrency: 1<<31 - 1,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	mux := asynq.NewServeMux()

	for taskType, task := range q.config.Tasks {
		mux.HandleFunc(taskType, task.taskHandler)
	}

	if err := srv.Run(mux); err != nil {
		panic(err)
	}
}

func (q *Queen) startClient() {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: q.config.RedisAddress, Password: q.config.RedisPassword, DB: q.config.RedisDB})
	defer func(client *asynq.Client) {
		err := client.Close()
		if err != nil {
			panic(err)
		}
	}(client)
	q.asynqClient = client
}

func (q *Queen) Enqueue(task *asynq.Task, options ...asynq.Option) (*asynq.TaskInfo, error) {
	return q.asynqClient.Enqueue(task, options...)
}

func (q *Queen) Ping() {

}
