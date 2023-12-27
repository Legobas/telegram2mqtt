package main

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

const (
	TIMEOUT    time.Duration = time.Second * 10
	SEND_TOPIC               = APPNAME + "/send"
)

type TelegramMessage struct {
	Message  string `json:"message"`
	ImageUrl string `json:"image-url"`
}

var mqttClient MQTT.Client

func sendToMtt(topic string, message string) {
	mqttClient.Publish(topic, byte(config.Mqtt.Qos), config.Mqtt.Retain, message)
}

func sendToMttRetain(topic string, message string) {
	mqttClient.Publish(topic, byte(config.Mqtt.Qos), true, message)
}

func receive(client MQTT.Client, msg MQTT.Message) {
	message := string(msg.Payload()[:])

	var telegramMessage TelegramMessage
	err := json.Unmarshal([]byte(message), &telegramMessage)
	if err != nil {
		log.Error().Msgf("JSON Error: %s", err.Error())
		return
	}

	err = validateMessage(telegramMessage)
	if err != nil {
		log.Warn().Msgf("MQTT message error: %s", err.Error())
		return
	}

	log.Debug().Msgf("Message: %s", telegramMessage.Message)

	sendToTelegram(telegramMessage.Message, "")
}

func GetClientId() string {
	hostname, _ := os.Hostname()
	return APPNAME + "_" + hostname
}

func validateMessage(msg TelegramMessage) error {
	if msg.Message == "" {
		return errors.New("message is mandatory")
	}

	return nil
}

func startMqttClient() {
	opts := MQTT.NewClientOptions().AddBroker(config.Mqtt.Url)
	if config.Mqtt.Username != "" && config.Mqtt.Password != "" {
		opts.SetUsername(config.Mqtt.Username)
		opts.SetPassword(config.Mqtt.Password)
	}
	opts.SetClientID(GetClientId())
	opts.SetCleanSession(true)
	opts.SetBinaryWill(APPNAME+"/status", []byte("Offline"), 0, true)
	opts.SetAutoReconnect(true)
	opts.SetConnectionLostHandler(connLostHandler)
	opts.SetOnConnectHandler(onConnectHandler)

	mqttClient = MQTT.NewClient(opts)
	token := mqttClient.Connect()
	if token.WaitTimeout(TIMEOUT) && token.Error() != nil {
		log.Fatal().Err(token.Error()).Msg("MQTT Connection Error")
	}

	token = mqttClient.Publish(APPNAME+"/status", 2, true, "Online")
	token.Wait()
}

func connLostHandler(c MQTT.Client, err error) {
	log.Fatal().Err(err).Msg("MQTT Connection lost")
}

func onConnectHandler(c MQTT.Client) {
	log.Debug().Msg("MQTT Client connected")
	token := mqttClient.Subscribe(SEND_TOPIC, 0, receive)
	if token.Wait() && token.Error() != nil {
		log.Fatal().Err(token.Error()).Msgf("Could not subscribe to %s", SEND_TOPIC)
	}
}
