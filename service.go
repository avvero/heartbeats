package main

import (
	"encoding/json"
	"errors"
	"time"
	"fmt"
	"log"
	"github.com/stretchr/stew/objects"
)

type Service struct {
	Url         string `json:"url,omitempty"`
	App         *Application `json:"app"`
	LastUpdated time.Time `json:"lastUpdated"`
	Error       string `json:"error,omitempty"`
	Original    string `json:"original,omitempty"`

	config *ServiceConfig
	parent *Service
	upd    chan time.Time
	events chan *Event

	MetricsUrl string `json:"metrics,omitempty"`
	metrics    chan *Metric

	Graphite string `json:"graphite,omitempty"`
}

type Application struct {
	Name        string `json:"name,omitempty"`
	Version     string `json:"version,omitempty"`
	Description string `json:"description,omitempty"`
	Tag         string `json:"tag,omitempty"`
	Components  []*Service `json:"components,omitempty"`
}

func (service *Service) update() {
	for {
		select {
		case t := <-service.upd:
			if service.App.Name != "" {
				log.Printf("Updating %s", service.App.Name)
			} else {
				log.Printf("Updating %s", service.Url)
			}
			for _, component := range service.App.Components {
				if component.upd != nil {
					component.upd <- t
				}
			}
			if service.config.Url != "" {
				body, err := callEndpoint(service.config.Url, service.config.Authorization)
				if err != nil {
					service.Error = fmt.Sprintf("%s", err)
					log.Printf("Error during updating %s: %s", service.config.Url, err)
				} else {
					//log.Printf("Response from %s: %s", service.config.Url, newService)
					service.Error = ""
					service.Original = string(body[:])
					err := service.merge(body)
					if err != nil {
						service.Error = fmt.Sprintf("%s", err)
						log.Printf("Error during updating %s: %s", service.config.Url, err)
					}
				}
			}
			service.LastUpdated = t
		}
	}
}
func (service *Service) merge(body []byte) error {
	// Merge from object
	newData := &Service{}
	json.Unmarshal(body, newData)
	if newData.App != nil {
		if service.config.Name == "" && newData.App.Name != "" {
			service.App.Name = newData.App.Name
		}
		if newData.App.Version != "" {
			if service.App.Version != "" &&
				newData.App.Version != "" &&
				service.App.Version != newData.App.Version {
				service.events <- &Event{
					service: service,
					description: fmt.Sprintf("%s has been updated from %s to %s", service.App.Name,
						service.App.Version, newData.App.Version)}
			}
			service.App.Version = newData.App.Version

		}
		if newData.App.Tag != "" {
			service.App.Tag = newData.App.Tag
		}
		if newData.App.Components != nil && len(newData.App.Components) > 0 {
			service.App.Components = newData.App.Components
		}
	}
	// If extension exists, use map
	if service.config.Extension != nil {
		var m map[string]interface{}
		json.Unmarshal(body, &m)
		if service.config.Extension.Name != "" && service.config.Name == "" {
			name := objects.Map(m).Get(service.config.Extension.Name)
			if name != nil {
				service.App.Name = name.(string)
			}
		}
		if service.config.Extension.Version != "" {
			version := objects.Map(m).Get(service.config.Extension.Version)
			if version != nil {
				newVersion := version.(string)
				if service.App.Version != "" && newVersion != "" && service.App.Version != newVersion {
					service.events <- &Event{
						service: service,
						description: fmt.Sprintf("%s has been updated from %s to %s",
							service.App.Name, service.App.Version, newVersion)}
				}
				service.App.Version = newVersion
			}
		}
		if service.config.Extension.Tag != "" {
			tag := objects.Map(m).Get(service.config.Extension.Tag)
			if tag != nil {
				service.App.Tag = tag.(string)
			}
		}

	}
	return nil
}

func (service *Service) findNearestNotificationConfig() (*Notification) {
	if service.config == nil {
		return nil
	}
	notification := service.config.Notification
	if notification != nil {
		return notification
	}
	if service.parent == nil {
		return nil
	}
	return service.parent.findNearestNotificationConfig()
}

func (service *Service) getFullAppName(it string) string {
	if service.parent == nil {
		return it
	}
	return service.parent.getFullAppName(fmt.Sprintf("%s.%s", service.parent.App.Name, it))
}

func (service *Service) sendEvents(eventConsumer EventConsumer) {
	for {
		select {
		case event := <-service.events:
			eventConsumer.consume(event)
		}
	}
}

func (service *Service) sendMetrics(metricConsumer MetricConsumer) {
	for {
		select {
		case metric := <-service.metrics:
			metricConsumer.consumeMetric(metric)
		}
	}
}

func (service *Service) runMetricsGrabber(pullInterval int) {
	ticker := time.NewTicker(time.Duration(pullInterval) * time.Second)
	go func() {
		for {
			select {
			case t := <-ticker.C:
				service.pullMetrics(t)
			}
		}
	}()
}
func (service *Service) pullMetrics(time time.Time) {
	body, err := callEndpoint(service.config.Metrics, service.config.Authorization)
	if err != nil {
		service.Error = fmt.Sprintf("%s", err)
		log.Printf("Error during metrics pulling %s: %s", service.config.Metrics, err)
		return
	}
	var m map[string]interface{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		log.Printf("Error during metrics mapping %s: %s", m, err)
		return
	}
	fullName := service.getFullAppName(service.App.Name)
	for k, v := range m {
		service.metrics <- &Metric{name: fmt.Sprintf("%s.%s", fullName, k), value: v, time: time}
	}
}

//Creates new service from config
func NewService(serviceConfig *ServiceConfig, parent *Service, eventConsumer EventConsumer,
	metricConsumer MetricConsumer, metricsPullInterval *int, graphiteDashboard *string) (*Service, error) {

	if serviceConfig.Url != "" && len(serviceConfig.Components) > 0 {
		return nil, errors.New("Service must not contain url and components in the same time")
	}
	service := &Service{
		App: &Application{
			Name: serviceConfig.Name,
			Tag:  serviceConfig.Tag,
		},
		Url:    serviceConfig.Url,
		config: serviceConfig,
		parent: parent,
		upd:    make(chan time.Time, 1000),
		events: make(chan *Event, 1000)}

	if serviceConfig.Metrics != "" {
		serviceFullName := service.getFullAppName(service.App.Name)
		service.MetricsUrl =  fmt.Sprintf("%s/?target=%s", *graphiteDashboard, serviceFullName)
		service.metrics = make(chan *Metric, 1000)
		go service.sendMetrics(metricConsumer)
		go service.runMetricsGrabber(*metricsPullInterval)
	}

	go service.update()
	go service.sendEvents(eventConsumer)

	if len(serviceConfig.Components) > 0 {
		service.App.Components = make([]*Service, len(serviceConfig.Components))
		for i, componentConfig := range serviceConfig.Components {
			s, err := NewService(componentConfig, service, eventConsumer, metricConsumer,
				metricsPullInterval, graphiteDashboard)
			if err != nil {
				return nil, err
			}
			service.App.Components[i] = s
		}
	}
	return service, nil
}
