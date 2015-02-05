package main

import (
	"fmt"
	"log"

	"github.com/bitly/go-nsq"
)

var (
	//TODO: build flags?
	numConcurrentProcesses = 3
	serverAddresses        = []string{"127.0.0.1:4161"}
)

type (
	ConsumerHandler struct {
		Topic   string
		Channel string
	}

	Topic struct {
		Name     string
		Channels []Channel
	}

	Channel struct {
		Name string
	}
)

func (h *ConsumerHandler) HandleMessage(msg *nsq.Message) error {
	if msg != nil {
		log.Printf("[%s/%s] %s", h.Topic, h.Channel, string(msg.Body))
	}
	return nil
}

func main() {
	var err error
	var consumer *nsq.Consumer
	var handler *ConsumerHandler

	config := nsq.NewConfig()
	consumers := make(map[string]*nsq.Consumer)

	//TODO: build flags?
	topics := []Topic{
		Topic{
			Name: "goapi",
			Channels: []Channel{
				Channel{Name: "metrics"},
			},
		},
	}

	for _, topic := range topics {
		for _, channel := range topic.Channels {
			if consumer, err = nsq.NewConsumer(topic.Name, channel.Name, config); err != nil {
				log.Println(err)
				continue
			}

			handler = &ConsumerHandler{
				Topic:   topic.Name,
				Channel: channel.Name,
			}

			consumer.AddConcurrentHandlers(handler, numConcurrentProcesses)

			if err = consumer.ConnectToNSQLookupds(serverAddresses); err != nil {
				consumer = nil
				log.Println(err)
				continue
			}

			consumers[fmt.Sprintf("%s/%s", topic.Name, channel.Name)] = consumer
		}
	}

	for {
		for _, c := range consumers {
			select {
			case <-c.StopChan:
				return
			}
		}
	}
}
