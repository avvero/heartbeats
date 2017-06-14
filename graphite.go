package main

import (
	"log"
	"fmt"
	"net"
	"time"
)

type Graphite struct {
	url              string
	metrics          chan *Metric
	herald           *Herald
	repeatSendOnFail bool
}

const (
	StopCharacter = "\r\n\r\n"
	SleepDuration = 5
)

func NewGraphite(url string, repeatSendOnFail bool) *Graphite {
	graphite := &Graphite{url: url, metrics: make(chan *Metric, 1000), repeatSendOnFail: repeatSendOnFail}
	go func() {
		for {
			select {
			case metric := <-graphite.metrics:
				times := 1
				for err := graphite.send(metric); err != nil && times < 4; {
					sleepTime := SleepDuration * time.Duration(times)
					log.Println(fmt.Sprintf("Will sleep for %v (%v attempt of 3)", sleepTime, times))
					time.Sleep(sleepTime * time.Second)
					times++
				}
			}
		}
	}()
	return graphite
}
func (graphite *Graphite) send(metric *Metric) error {
	message := fmt.Sprintf("%s %v %v", metric.name, metric.value, metric.time.Unix())

	conn, err := net.Dial("tcp", graphite.url)
	if err != nil {
		if graphite.repeatSendOnFail {
			log.Println(fmt.Sprintf("Could not connected to graphite: %v", graphite.url))
			return err
		} else {
			log.Println(fmt.Sprintf("Could not connected to graphite: %v, metrics will be loosed", graphite.url))
			return nil
		}
	}
	defer conn.Close()

	conn.Write([]byte(message))
	conn.Write([]byte(StopCharacter))
	log.Println(fmt.Sprintf("Send metrics: %s %s", graphite.url, message))
	return nil
}

func (graphite *Graphite) consumeMetric(metric *Metric) {
	graphite.metrics <- metric
}
