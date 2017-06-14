package main

import (
	_ "net/http/pprof"
	"time"
	"net/http"
	"encoding/json"
	"fmt"
	"log"
	"flag"
)

var (
	httpPort                 = flag.String("httpPort", "8080", "http server port")
	updateInterval           = flag.Int("infoUpdateInterval", 5, "update interval for infos")
	heraldEndpoint           = flag.String("heraldEndpoint", "", "endoint to send messages to herald the bot")
	metricsPullInterval      = flag.Int("metricsPullInterval", 5, "pull interval for metrics")
	graphiteUrl              = flag.String("graphiteUrl", "", "host and port to send plaint text metrics to graphite")
	graphiteDashboard        = flag.String("graphiteDashboard", "", "url for graphite dashboard")
	graphiteRepeatSendOnFail = flag.Bool("graphiteRepeatSendOnFail", false, "repeat send metrcis to graphite on fail")
)

func main() {
	flag.Parse()

	var eventConsumer EventConsumer
	if *heraldEndpoint != "" {
		eventConsumer = NewHerald(*heraldEndpoint)
		log.Println("Events will be passed to herald: " + *heraldEndpoint)
	} else {
		eventConsumer = &EventConsumerStub{}
	}
	log.Println(fmt.Sprintf("eventConsumer %v", eventConsumer))

	var metricConsumer MetricConsumer
	if *graphiteUrl != "" {
		metricConsumer = NewGraphite(*graphiteUrl, *graphiteRepeatSendOnFail)
		log.Println("Metrics will be passed to graphite: " + *heraldEndpoint)
	} else {
		metricConsumer = &MetricConsumerStub{}
	}

	rootConfig, err := ReadConfig("services.json")
	if err != nil {
		panic(fmt.Sprintf("Error during configuration %v", err))
	}

	rootService, err := NewService(rootConfig, nil, eventConsumer, metricConsumer, metricsPullInterval,
		graphiteDashboard)
	if err != nil {
		panic(fmt.Sprintf("Error during configuration %v", err))
	}

	ticker := time.NewTicker(time.Duration(*updateInterval) * time.Second)
	rootService.upd <- time.Now()
	go func() {
		for {
			select {
			case <-ticker.C:
				rootService.upd <- time.Now()
			}
		}
	}()

	// proxy stuff
	http.Handle("/", http.FileServer(http.Dir("public")))
	http.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		js, err := json.Marshal(rootService)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	})

	log.Println("Http server started on port " + *httpPort)
	http.ListenAndServe(":" + *httpPort, nil)
}
