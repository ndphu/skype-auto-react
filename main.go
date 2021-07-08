package main

import (
	"github.com/ndphu/message-handler-lib/broker"
	"github.com/ndphu/message-handler-lib/config"
	"github.com/ndphu/message-handler-lib/handler"
	"github.com/ndphu/message-handler-lib/model"
	"github.com/ndphu/skype-auto-react/rule"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	if err := rule.LoadRules(); err != nil {
		log.Fatalf("Fail to load rules config by error %v\n", err)
	}

	workerId, consumerId := config.LoadConfig()
	eventHandler, err := handler.NewEventHandler(handler.EventHandlerConfig{
		WorkerId:            workerId,
		ConsumerId:          consumerId,
		ConsumerWorkerCount: 8,
		ServiceName:         "auto-react",
	}, func(e model.MessageEvent) {
		processMessage(e)
	})

	if err != nil {
		log.Fatalf("Fail to create handler by error %v\n", err)
	}

	eventHandler.Start()

	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	<-termChan
	log.Println("Shutdown signal received")
	eventHandler.Stop()
}

func processMessage(evt model.MessageEvent) {
	log.Println("Received message", evt.MessageId(), "from", evt.GetFrom(), "in conversation", evt.GetThreadId())
	rules := rule.GetRules(evt.GetFrom(), evt.GetThreadId())
	workerId := os.Getenv("WORKER_ID")
	for _, r := range rules {
		for _, reaction := range r.Reacts {
			go react(workerId, evt, reaction)
		}
	}
}

func react(workerId string, evt model.MessageEvent, react string) error {
	log.Println("React: Reacting message", evt.MessageId(), "with", react)
	rpc, err := broker.NewRpcClient(workerId)
	if err != nil {
		log.Println("React: Fail to create RPC client")
		return err
	}

	request := NewReactRequest(evt.GetThreadId(), evt.MessageId(), react)
	if err := rpc.Send(request); err != nil {
		log.Println("React: Fail to react message", err.Error())
		return err
	}

	log.Println("React: Successfully react message", react)
	return nil
}

func NewReactRequest(threadId, messageId, react string) *broker.RpcRequest {
	return &broker.RpcRequest{
		Method: "react",
		Args:   []string{threadId, messageId, react},
	}
}
