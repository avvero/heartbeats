package main

import (
	"os"
	"encoding/json"
)

type Authorization struct {
	Login string
	Pass  string
}

type Extension struct {
	Name    string
	Version string
	Tag     string
}

type Notification struct {
	Conversations []string
}

type ServiceConfig struct {
	Url     string
	Metrics string
	Name    string
	Version string
	Tag     string

	Components    []*ServiceConfig
	Extension     *Extension
	Authorization *Authorization
	Notification  *Notification
}

// Read config
func ReadConfig(fileName string) (*ServiceConfig, error) {
	rootConfig := ServiceConfig{}
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&rootConfig)
	if err != nil {
		return nil, err
	}
	return &rootConfig, nil
}
