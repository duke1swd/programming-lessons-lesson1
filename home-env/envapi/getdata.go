/*
 * Get the environmental data from mqtt.
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
)

var (
	string tempTopic
	string luxTopic
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
}

// Call the cancel function after a deadline.  Each time any value is received
// on the channel, the deadline is extended.
func timeoutRoutine(c context.Context, cf context.CancelFunc, d time.Duration, ch chan int) {
	for {
		subContext, cfl := context.WithTimeout(c, d)
		select {
		case <-subContext.Done():
			cfl()
			cf()
			return
		case <-ch:
			cfl()
		}
	}
}

// This routine is called when we receive a message.
// If it is one we are interested in, we save the payload.
var f1 mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := string(msg.Payload())

	// Ignore broadcast messages
	if !flagpPresent && strings.Contains(topic, "$broadcast") {
		return
	}

	switch payload {
	case tempTopic:
		temp = payload;
	case luxTopic:
		lux = payload;
	default:
	}

	// tell the world we are still working
	timeoutChannel <- 0
}

func getClient() {
	opts := mqtt.NewClientOptions().AddBroker("tcp://127.0.0.1:1883").SetClientID("env-api-server")
	opts.SetKeepAlive(60 * time.Second)
	opts.SetDefaultPublishHandler(f1)
	opts.SetPingTimeout(1 * time.Second)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	mqttClient = c
}

// get all the persistent messages and build a map of everything we know about everybody
func getDevices() {
	c := mqttClient

	subscription := "devices/#"
	if flagpPresent {
		subscription = "#"
	}

	if token := c.Subscribe(subscription, 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	// Set up to wait for 1 second after last message is received
	timeoutContext, cf := context.WithCancel(context.Background())
	go timeoutRoutine(timeoutContext, cf, time.Second, timeoutChannel)
	select {
	case <-timeoutContext.Done():
	}

	if token := c.Unsubscribe("devices/#"); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
}

// List all devices
func deviceInfo() {
	keys := make([]string, len(deviceMap))

	i := 0
	maxLen := 0
	for s := range deviceMap {
		keys[i] = s
		i++
		if len(s) > maxLen {
			maxLen = len(s)
		}
	}
	// now, sort the keys
	sort.Strings(keys)

	// print things out in alpha order
	for _, dev := range keys {
		properties := deviceMap[dev]

		state, ok := properties["$state"]
		if !ok {
			state = "<unknown>"
		}
		if flagL {
			i = maxLen - len(dev)
			for  i > 0 {
				fmt.Printf(" ")
				i--
			}
			fmt.Printf("%s: state=%s\n", dev, state)
		} else if state == "lost" || state == "disconnected" {
			fmt.Println(dev)
		}
	}
}

// returns true on error
func clearDevice(device string) bool {
	properties, ok := deviceMap[device]

	if !ok {
		fmt.Printf("Device %s not found\n", device)
		return true
	}

	state := properties["$state"]
	if !flagf && state != "lost" && state != "disconnected" {
		fmt.Printf("Device %s not known to be offline\n", device)
		return true
	}
	fmt.Printf("Clearing device %s\n", device)

	for topic, _ := range properties {
		topic = "devices/" + device + "/" + topic
		if flagD {
			fmt.Printf("\t%s\n", topic)
		}
		publishToken := mqttClient.Publish(topic, 0, true, "")
		if publishToken.Wait() && publishToken.Error() != nil {
			panic(publishToken.Error())
		}
	}
	return false
}

/*
 * Create a slice of strings.  Each entry is a topic that matched.
 * Sort the slice in alpha order
 */
func topicMatch(topic string) bool {
	return strings.HasPrefix(topic, flagp)
}

func buildTopics() {

	for k, _ := range allTopics {
		if topicMatch(k) {
			selectedTopics = append(selectedTopics, k)
		}
	}

	// sort the slice
	sort.Slice(selectedTopics, func(i, j int) bool { return selectedTopics[i] < selectedTopics[j] })
}

func topicInfo() {
	l := 0
	for _, v := range selectedTopics {
		if len(v) > l {
			l = len(v)
		}
	}

	f := "%" + strconv.Itoa(l) + "s: %s\n"
	for _, v := range selectedTopics {
		fmt.Printf(f, v, allTopics[v])
	}
}

func topicClear() {
	fmt.Println("Clearing topics:")
	for _, v := range selectedTopics {
		fmt.Printf("\t%s\n", v)
		publishToken := mqttClient.Publish(v, 0, true, "")
		if publishToken.Wait() && publishToken.Error() != nil {
			panic(publishToken.Error())
		}
	}
}

func main() {
	var errors bool = false

	//mqtt.DEBUG = log.New(os.Stdout, "", 0)
	mqtt.ERROR = log.New(os.Stdout, "", 0)

	getClient()
	getDevices()

	if flagpPresent {
		buildTopics()
		if flagl {
			topicInfo()
		} else if flagf {
			topicClear()
		}
	} else {

		if flagl || flagL {
			deviceInfo()
		}

		if flagcPresent {
			if flagc == "ALL" {
				for d, p := range deviceMap {
					state := p["$state"]
					if state == "lost" || state == "disconnected" {
						if clearDevice(d) {
							errors = true
						}
					}
				}
			} else {
				errors = clearDevice(flagc)
			}
			if errors {
				os.Exit(1)
			}
		}

	}
}
