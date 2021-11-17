package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"
	tb "gopkg.in/tucnak/telebot.v2"
	"net/http"
	"strconv"
)

var (
	TelebotSecretErr = "TELEBOT_SECRET is undefined"
	TelebotSecret    = "" //os.Getenv("TELEBOT_SECRET")
)

func handler(req events.APIGatewayProxyRequest) (resp events.APIGatewayProxyResponse, err error) {
	var chatId int64
	resp = events.APIGatewayProxyResponse{StatusCode: http.StatusOK}

	if TelebotSecret == "" {
		err = errors.New(TelebotSecretErr)
		resp.Body = TelebotSecretErr
		return
	}

	//var payload interface{}
	//err = json.Unmarshal([]byte(req.Body), &payload)
	//if err != nil {
	//	println(err.Error() + " body:" + req.Body)
	//	return
	//}

	// https://github.com/onemorebit/devch2021_bot/actions/runs/{id}
	headerTgChatID, ok := req.Headers["x-tg-chat-id"]
	if !ok {
		resp.Body = "missing chat id"
		return
	}
	chatId, err = strconv.ParseInt(headerTgChatID, 10, 64)
	if err != nil {
		resp.Body = err.Error()
		return
	}

	b, err := tb.NewBot(tb.Settings{
		Token:       TelebotSecret,
		Synchronous: true,
	})

	if err != nil {
		resp.Body = err.Error()
		return
	}

	_, err = b.Send(&tb.Chat{ID: chatId}, req.Body, &tb.SendOptions{ParseMode: tb.ModeMarkdownV2, DisableWebPagePreview: true})
	if err != nil {
		resp.Body = err.Error()
	}
	return
}

func main() {
	lambda.Start(handler)
}
