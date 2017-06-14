package main

import "time"

type Metric struct {
	name  string
	time  time.Time
	value interface{}
}

type MetricConsumer interface {
	consumeMetric(metric *Metric)
}

type MetricConsumerStub struct {
}

func (it *MetricConsumerStub) consumeMetric(event *Metric) {
	//do nothing
}

