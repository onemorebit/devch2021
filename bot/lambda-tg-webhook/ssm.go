package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	tb "gopkg.in/tucnak/telebot.v2"
	"strings"
	"time"
)

func SSMRunCommand(shellCmds []string) string {
	ssmSes := ssm.New(sess)
	vmid, err := getVmID()
	if err != nil {
		return err.Error()
	}

	sendCmd, err := ssmSes.SendCommand(&ssm.SendCommandInput{
		InstanceIds:  aws.StringSlice([]string{vmid}),
		DocumentName: aws.String("AWS-RunShellScript"),
		Comment:      aws.String("triggered through lambda."),
		Parameters: map[string][]*string{
			"commands": aws.StringSlice(shellCmds),
		},
	})
	if err != nil {
		return "Send command err: " + err.Error()
	}
	time.Sleep(2 * time.Second)
	cmdInvocation, err := ssmSes.GetCommandInvocation(&ssm.GetCommandInvocationInput{
		CommandId:  sendCmd.Command.CommandId,
		InstanceId: aws.String(vmid),
	})

	switch {
	case err != nil:
		return fmt.Sprintf(
			"GetCommandInvocation err:\nCommandId: %s; InstanceId: %s;\n\n%s",
			*sendCmd.Command.CommandId, vmid, err.Error())
	case cmdInvocation != nil:
		return fmt.Sprintf("Executed:\n```\n%v\n```\nResult:\n%s", strings.Join(shellCmds[:], "\n"), *cmdInvocation.StandardErrorContent+*cmdInvocation.StandardOutputContent)
	}
	return "unexpected behavior"

}

func tgShowVerVM(b *tb.Bot, m *tb.Message) {
	if !ifStackCreated() {
		b.Send(m.Chat, "VM stack is not exist")
		return
	}
	cmd := []string{
		"PATH=/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin:/root/bin",
		"kf5-config --version 2>/dev/null",
		"uname -a && date -u",
		"echo -n Private ip:",
		"curl -s 169.254.169.254/latest/meta-data/local-ipv4",
		"echo",
		"echo -n Public ip:",
		"curl -s 169.254.169.254/latest/meta-data/public-ipv4",
	}
	ssmStdOut := SSMRunCommand(cmd)

	b.Send(m.Chat, ssmStdOut)

}

func tgKdePatchVM(b *tb.Bot, m *tb.Message) {
	if !ifStackCreated() {
		b.Send(m.Chat, "VM stack is not exist")
		return
	}
	cmd := []string{
		"PATH=/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin:/root/bin",
		"pkg delete -y postgresql12-client",
		"mkdir -p /usr/local/etc/pkg/repos",
		"cp /etc/pkg/FreeBSD.conf /usr/local/etc/pkg/repos/",
		"sed -I bak s/quarterly/latest/ /usr/local/etc/pkg/repos/FreeBSD.conf",
		"pkg update",
		"pkg upgrade -y",
		"echo patched",
	}
	ssmStdOut := SSMRunCommand(cmd)

	b.Send(m.Chat, fmt.Sprintf("%s\n\nPlease run %s again", ssmStdOut, TbCmdOnVmShowVer))
}

func tgMonitorVM(b *tb.Bot, m *tb.Message) {
	if !ifStackCreated() {
		b.Send(m.Chat, "VM stack is not exist")
		return
	}
	cmd := []string{
		"PATH=/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin:/root/bin",
		"top -b -z -w -n 10",
	}
	ssmStdOut := SSMRunCommand(cmd)
	b.Send(m.Chat, ssmStdOut)
}
