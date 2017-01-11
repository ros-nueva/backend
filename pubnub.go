package main

import (
	"github.com/pubnub/go/messaging"
	"encoding/json"
	"log"
)

const (
	PubKey = "pub-c-5277d936-a028-41b2-8774-f7f7cb8d2102"
	SubKey = "sub-c-056c9aec-d84b-11e6-b9cf-02ee2ddab7fe"
	SecKey = "sec-c-NTVmYWEyYjYtMzU5OC00YTJlLTgwNjAtNTdiMzlmOGU1Nzk1"
	CipKey = ""
	SSL    = false
	UUID   = ""
	Channel = "unicub"
)

type PubnubManager struct {
	*messaging.Pubnub
}

type MessageStart struct {
	UserID string `json:"user_id"`
	JourneyID string `json:"journey_id"`
}

func (manager *PubnubManager) Initialize() {
	manager.Pubnub = messaging.NewPubnub(PubKey, SubKey, SecKey, CipKey, SSL, UUID)
}

func (manager *PubnubManager) PublishJSON(message interface{}) {
	jsonMsg, _ := json.Marshal(message)
	successChannel := make(chan []byte)
	errorChannel := make(chan []byte)
	go manager.Publish(Channel, jsonMsg, successChannel, errorChannel)
	select {
	case response := <-successChannel:
		log.Println(string(response))
	case err := <-errorChannel:
		log.Println(string(err))
	case <-messaging.Timeout():
		log.Println("Publish() timeout")
	}
}