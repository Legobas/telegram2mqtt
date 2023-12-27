package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	CONFIG_FILE = "telegram2mqtt.yml"
	CONFIG_DIR  = ".config"
	CONFIG_ROOT = "/config"
)

type Mqtt struct {
	Url      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Qos      int    `yaml:"qos"`
	Retain   bool   `yaml:"retain"`
}

type Telegram struct {
	ApiKey string `yaml:"api_key"`
	ChatId string `yaml:"chat_id"`
}

type Keys struct {
	Title       string `yaml:"title"`
	Row         int    `yaml:"row"`
	MqttTopic   string `yaml:"mqtt-topic"`
	MqttMessage string `yaml:"mqtt-message"`
}

type Command struct {
	Command      string `yaml:"command"`
	ReplyMessage string `yaml:"replymessage"`
	Keyboards    []Keys `yaml:"keyboard"`
	MqttTopic    string `yaml:"mqtt-topic"`
	MqttMessage  string `yaml:"mqtt-message"`
}

type Config struct {
	Mqtt     Mqtt      `yaml:"mqtt"`
	Telegram Telegram  `yaml:"telegram"`
	Commands []Command `yaml:"commands"`
}

func getConfig() Config {
	var config Config

	configFile := filepath.Join(CONFIG_ROOT, CONFIG_FILE)
	msg := configFile
	data, err := os.ReadFile(configFile)
	if err != nil {
		homedir, _ := os.UserHomeDir()
		configFile := filepath.Join(homedir, CONFIG_DIR, CONFIG_FILE)
		msg += ", " + configFile
		data, err = os.ReadFile(configFile)
	}
	if err != nil {
		workingdir, _ := os.Getwd()
		configFile := filepath.Join(workingdir, CONFIG_FILE)
		msg += ", " + configFile
		data, err = os.ReadFile(configFile)
	}
	if err != nil {
		msg = "Configuration file could not be found: " + msg
		log.Fatal(msg)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatal(err)
	}

	err = validate(config)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v\n", config)
	return config
}

func validate(config Config) error {
	if config.Mqtt.Url == "" {
		return errors.New("Config error: MQTT Server URL is mandatory")
	}
	if config.Telegram.ApiKey == "" {
		return errors.New("Config error: Telegram API Key is mandatory")
	}
	if len(config.Commands) == 0 {
		return errors.New("Config error: No Commands defined")
	}

	return nil
}
