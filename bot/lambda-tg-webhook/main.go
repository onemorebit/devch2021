package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"
	tb "gopkg.in/tucnak/telebot.v2"
	"net/http"
	"strconv"
	"strings"
)

var (
	TelebotSecret     = "" //os.Getenv("TELEBOT_SECRET")
	TbCmdDeployApp    = "/deploy"
	TbCmdOnVmCreate   = "/vm_create"
	TbCmdOnVmShowVer  = "/vm_show_version"
	TbCmdOnVmKdePatch = "/vm_kde_patch"
	TbCmdOnVmCDestroy = "/vm_destroy"
	TbCmdOnVmMonitor  = "/vm_monitor"
	DefaultText       = "Supported commands:\n" +
		TbCmdDeployApp + " deploy app\n" +
		TbCmdOnVmCreate + " create freebsd with the old KDE version\n" +
		TbCmdOnVmShowVer + " show FreeBsd and KDE versions\n" +
		TbCmdOnVmKdePatch + " patch KDE\n" +
		TbCmdOnVmCDestroy + " destroy the FreeBsd instance\n" +
		TbCmdOnVmMonitor + " get instance stats\n"

	DeployBranch = "main"
	DeployURL    = "" //os.Getenv("DEPLOY_URL")
)

func handler(req events.APIGatewayProxyRequest) (resp events.APIGatewayProxyResponse, err error) {
	resp = events.APIGatewayProxyResponse{StatusCode: http.StatusOK}

	if TelebotSecret == "" { // see Makefile
		err = errors.New("TELEBOT_SECRET is undefined")
		return
	}
	callbackUrl := fmt.Sprintf("https://%s/%s/callback-curl",
		req.Headers["Host"],
		req.RequestContext.Stage,
	)
	var u tb.Update
	err = json.Unmarshal([]byte(req.Body), &u)
	if err != nil {
		println(err.Error() + " body:" + req.Body)
		return
	}
	b, err := tb.NewBot(tb.Settings{
		Token:       TelebotSecret,
		Synchronous: true,
	})

	if err != nil {
		return
	}
	b.Handle(TbCmdDeployApp, func(m *tb.Message) {
		var err error
		b.Reply(m, "starting the deployment, msg id: "+strconv.Itoa(m.ID))
		ghPayload := fmt.Sprintf("{%q: %q, %q: {%q: %q, %q: %q}}",
			"ref", DeployBranch,
			"inputs",
			"callbackUrl", callbackUrl,
			"chatId", strconv.FormatInt(m.Chat.ID, 10),
		)
		ghResp, err := http.Post(
			DeployURL,
			"application/json",
			strings.NewReader(ghPayload),
		)

		if err != nil {
			b.Send(m.Chat, "error:\n"+err.Error())
			return
		}
		defer ghResp.Body.Close()
		if ghResp.StatusCode != http.StatusNoContent {
			// debug resp
			var headrz bytes.Buffer

			headrz.WriteString(ghResp.Status)
			headrz.WriteString("\n")
			headrz.WriteString(DeployURL[:10])
			headrz.WriteString("\n")
			for k, v := range ghResp.Header {
				headrz.WriteString(k)
				headrz.WriteString(" : ")
				headrz.WriteString(fmt.Sprint(v))
				headrz.WriteString("\n")
			}
			body := make([]byte, 192)
			ghResp.Body.Read(body)
			headrz.Write(body)

			println("unexpected behavior:" + headrz.String())
			b.Send(m.Chat, "something went wrong. see trail logs for more details.")

		}
	})
	b.Handle(TbCmdOnVmCreate, func(m *tb.Message) { tgCreateVM(b, m) })
	b.Handle(TbCmdOnVmShowVer, func(m *tb.Message) { b.Send(m.Chat, TbCmdOnVmShowVer+" is not implemented yet") })
	b.Handle(TbCmdOnVmKdePatch, func(m *tb.Message) { b.Send(m.Chat, TbCmdOnVmKdePatch+" is not implemented yet") })
	b.Handle(TbCmdOnVmCDestroy, func(m *tb.Message) { b.Send(m.Chat, TbCmdOnVmCDestroy+" is not implemented yet") })
	b.Handle(TbCmdOnVmMonitor, func(m *tb.Message) { b.Send(m.Chat, TbCmdOnVmMonitor+" is not implemented yet") })
	b.Handle(tb.OnText, func(m *tb.Message) { b.Send(m.Chat, DefaultText) })
	b.Handle(tb.OnUserJoined, func(m *tb.Message) { b.Send(m.Chat, DefaultText) })
	println("processing message: " + strconv.Itoa(u.Message.ID))
	b.ProcessUpdate(u)
	return
}

func main() {
	lambda.Start(handler)
}
