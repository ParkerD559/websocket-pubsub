package main

import (
	"encoding/json"
	"fmt"

	"github.com/alash3al/go-pubsub"
	uuid "github.com/nu7hatch/gouuid"

	melody "gopkg.in/olahol/melody.v1"
)

var subscribers pubsub.Subscribers
var pubsubBroker *pubsub.Broker

type messageAggStruct struct {
	Msg       *pubsub.Message
	WsSession *melody.Session
}

var aggChannel chan messageAggStruct

func init() {
	pubsubBroker = pubsub.NewBroker()
	aggChannel = make(chan messageAggStruct)
	subscribers = make(pubsub.Subscribers)
	go emitListener()
}

func emitListener() {
	// Listen on the aggregation channel, send what we get
	for {
		select {
		case message := <-aggChannel:
			if message.Msg == nil {
				continue
			}
			wsMsg := wsMessage{
				Topic: message.Msg.GetTopic(),
				Data:  message.Msg.GetPayload(),
			}
			msg, _ := json.Marshal(wsMsg)
			message.WsSession.Write(msg)
		}
	}
}

func addListener(topic string, listener *melody.Session) {
	ID, _ := listener.Get("connID")
	userID := ID.(*uuid.UUID).String()
	subscriber, subExists := subscribers[userID]
	fmt.Println(subscribers)
	// Create subscriber if it does not yet exist
	if !subExists {
		subscriber, _ = pubsubBroker.Attach()
		subscribers[userID] = subscriber
		// Send all messages to one channel
		go func() {
			for {
				select {
				case msg := <-subscriber.GetMessages():
					aggChannel <- messageAggStruct{msg, listener}
				}
			}
		}()
	}
	pubsubBroker.Subscribe(subscriber, topic)
}

func removeListener(topic string, listener *melody.Session) {
	ID, _ := listener.Get("connID")
	userID := ID.(*uuid.UUID).String()
	subscriber, ok := subscribers[userID]
	if ok {
		pubsubBroker.Unsubscribe(subscriber, topic)
	}
}

func disconnectListener(listener *melody.Session) {
	ID, _ := listener.Get("connID")
	userID := ID.(*uuid.UUID).String()
	subscriber, ok := subscribers[userID]
	if ok {
		pubsubBroker.Detach(subscriber)
		delete(subscribers, userID)
	}
}

func publish(topic string, payload interface{}) {
	pubsubBroker.Broadcast(payload, topic)
}
