package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

const (
	TOPIC = APPNAME + "/message"
)

var telegramBot *bot.Bot
var ctx context.Context

func StartTelegramClient() {
	var cancel context.CancelFunc

	ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	var err error
	telegramBot, err = bot.New(config.Telegram.ApiKey, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("Telegram client error")
	}

	telegramBot.Start(ctx)
}

func sendToTelegram(message string, imageUrl string) {
	telegramBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: config.Telegram.ChatId,
		Text:   message,
	})
}

// handler
func handler(ctx context.Context, tBot *bot.Bot, update *models.Update) {

	if update.Message.Text == "/info" {
		userInfo := update.Message.From.Username
		if update.Message.From.FirstName != "" || update.Message.From.LastName != "" {
			userInfo += " ("
			if update.Message.From.FirstName != "" {
				userInfo += update.Message.From.FirstName
			}
			if update.Message.From.LastName != "" {
				userInfo += " " + update.Message.From.LastName
			}
			userInfo += ")"
		}
		userInfo += " - " + strings.ToUpper(update.Message.From.LanguageCode)
		if update.Message.From.IsBot {
			userInfo += " [BOT]"
		}
		botInfo := ""
		botname, err := tBot.GetMyName(ctx, &bot.GetMyNameParams{})
		if err == nil {
			botInfo += botname.Name
		} else {
			log.Error().Msgf("Error: %s", err.Error())
		}
		sdescr, err := tBot.GetMyShortDescription(ctx, &bot.GetMyShortDescriptionParams{})
		if err == nil {
			if sdescr.ShortDescription != "" {
				botInfo += " - " + sdescr.ShortDescription
			}
		} else {
			log.Error().Msgf("Error: %s", err.Error())
		}
		descr, err := tBot.GetMyDescription(ctx, &bot.GetMyDescriptionParams{})
		if err == nil {
			if descr.Description != "" {
				botInfo += " (" + descr.Description + ")"
			}
		} else {
			log.Error().Msgf("Error: %s", err.Error())
		}
		botName := "@"
		me, err := tBot.GetMe(ctx)
		if err == nil {
			if me.Username != "" {
				botName += me.Username
			}
		} else {
			log.Error().Msgf("Error: %s", err.Error())
		}

		replyMessage := APPNAME + " " + VERSION + "\n"
		replyMessage += "User: " + userInfo + "\n"
		replyMessage += "Chat: " + update.Message.Chat.Type + "\n"
		replyMessage += "Bot " + botName + ": " + botInfo + "\n"
		chatId := ""
		if config.Telegram.ChatId == "" {
			chatId = strconv.FormatInt(update.Message.Chat.ID, 10)
			replyMessage += "Your Chat ID is: " + chatId + "\n"
		} else {
			chatId = config.Telegram.ChatId
		}

		log.Info().Msg(replyMessage)

		tBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatId,
			Text:   replyMessage,
		})
	}

	// No other commands available if config has no ChatId
	if config.Telegram.ChatId == "" {
		return
	}

	msgs := strings.Split(update.Message.Text, " ")
	for _, cmd := range config.Commands {
		if msgs[0] == cmd.Command {
			log.Debug().Msgf("Command: %s", cmd.Command)
			if cmd.MqttMessage == "" && len(msgs) > 1 {
				log.Debug().Msgf("Topic: %s, Msg: %s", cmd.MqttTopic, msgs[1])
				sendToMtt(cmd.MqttTopic, msgs[1])
			} else {
				log.Debug().Msgf("MqttTopic: %s, MqttMessage: %s", cmd.MqttTopic, cmd.MqttMessage)
				sendToMtt(cmd.MqttTopic, cmd.MqttMessage)
			}
			return
		}
	}
}
