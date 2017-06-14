package main

import (
	"encoding/json"
	"bytes"
	"net/http"
	"io"
	"os"
	"log"
	"fmt"
)

type Message struct {
	ConversationId string `json:"conversationId"`
	Text           string `json:"text"`
}

type Herald struct {
	url      string
	messages chan *Message
	events   chan *Event
}

func NewHerald(url string) *Herald {
	herald := &Herald{url: url, messages: make(chan *Message), events: make(chan *Event)}
	go herald.doSend()
	go herald.doConsume()
	return herald
}

func (herald *Herald) doSend() {
	for {
		select {
		case message := <-herald.messages:
			log.Println(fmt.Sprintf("Send message to herald: %s", *message))
			b := new(bytes.Buffer)
			json.NewEncoder(b).Encode(message)
			res, _ := http.Post(herald.url, "application/json; charset=utf-8", b)
			io.Copy(os.Stdout, res.Body)
		}
	}
}

func (herald *Herald) doConsume() {
	for {
		select {
		case event := <-herald.events:
			notification := event.service.findNearestNotificationConfig()
			if notification != nil {
				for _, conversation := range notification.Conversations {
					text := event.description
					if event.service.parent != nil {
						text = fmt.Sprintf("%s: %s", event.service.parent.App.Name,
							event.description)
					}
					herald.messages <- &Message{ConversationId: conversation, Text: text}
				}
			}
		}
	}
}

func (herald *Herald) consume(event *Event) {
	herald.events <- event
}
