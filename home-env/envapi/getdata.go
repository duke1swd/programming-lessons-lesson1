/*
 * Get the environmental data from mqtt.
 */

package main

import (
	"time"
	"strings"
	"fmt"
	"os"

	"github.com/eclipse/paho.mqtt.golang"
)

var (
	string tempTopic
	string luxTopic
	string mqttAddress
	string mqttClientId
	string mqttSubscription
)

var (
	mqttClient          mqtt.Client
	epoch               time.Time
	timeoutContext      context.Context
	timeoutChannel      chan int = make(chan int)
)

func init() {
	epoch, _ = time.Parse("2006-Jan-02 MST", "2018-Nov-01 EDT")
	tempTopic = "environment/outdoor-temp"
	luxTopic = "environment/outdoor-lux"
	mqttAddress = "tcp://127.0.0.1:1883"
	mqttClientId = "env-api-server"
	mqttSubscription = "environment/#"
}

// This routine is called when we receive a message.
// If it is one we are interested in, we save the payload.
var f1 mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := string(msg.Payload())

	// Ignore broadcast messages
	if strings.Contains(topic, "$broadcast") {
		return
	}

	switch payload {
	case tempTopic:
		temp = payload;
	case luxTopic:
		lux = payload;
	default:
	}
}

func getClient() {
	opts := mqtt.NewClientOptions().AddBroker(mqttAddress).SetClientID(mqttClientId)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetDefaultPublishHandler(f1)
	opts.SetPingTimeout(1 * time.Second)

	c := mqtt.NewClient(opts)
	c.SetOrderMatters(false)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	if token := c.Subscribe(mqttSubscription, 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	mqttClient = c
}
