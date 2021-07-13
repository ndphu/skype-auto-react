package main

import (
	"fmt"
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
	serviceName := "skype-auto-react"

	textHandler, err := handler.NewEventHandler(handler.EventHandlerConfig{
		WorkerId:            workerId,
		ConsumerId:          consumerId,
		ConsumerWorkerCount: 8,
		ServiceName:         serviceName,
		QueueNameOverrideCallback: func() (string, string) {
			textExchange := "/worker/" + workerId + "/textMessages"
			textQueueName := "/message-handler/" + serviceName + "/text-queue-" + consumerId
			return textExchange, textQueueName
		},
		RemoveQueue: true,
	}, func(e model.MessageEvent) {
		processMessage(e)
	})
	if err != nil {
		log.Fatalf("Fail to create text message event handler by error=%v\n", err)
	}

	textHandler.Start()

	mediaHandler, err := handler.NewEventHandler(handler.EventHandlerConfig{
		WorkerId:            workerId,
		ConsumerId:          consumerId,
		ConsumerWorkerCount: 8,
		ServiceName:         serviceName,
		QueueNameOverrideCallback: func() (string, string) {
			textExchange := "/worker/" + workerId + "/mediaMessages"
			textQueueName := "/message-handler/" + serviceName + "/media-queue-" + consumerId
			return textExchange, textQueueName
		},
		RemoveQueue: true,
	}, func(e model.MessageEvent) {
		processMessage(e)
	})
	if err != nil {
		log.Fatalf("Fail to create media message event handler by error=%v\n", err)
	}

	mediaHandler.Start()

	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	<-termChan
	log.Println("Shutdown signal received")
	textHandler.Stop()
	mediaHandler.Stop()
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
		log.Printf("React: Fail to create RPC client by error %v\n", err)
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
func wrapAsPreformatted(message string) string {
	return fmt.Sprintf("<pre raw_pre=\"{code}\" raw_post=\"{code}\">%s</pre>", message)
}
