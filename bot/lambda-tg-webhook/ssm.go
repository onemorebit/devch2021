package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	tb "gopkg.in/tucnak/telebot.v2"
	"strings"
	"time"
)

func curentCmdInProgress(ssmSes *ssm.SSM, vmid string) bool {

	cmdInvocation, err := ssmSes.ListCommands(&ssm.ListCommandsInput{
		Filters: []*ssm.CommandFilter{
			{Key: aws.String("DocumentName"), Value: aws.String("AWS-RunShellScript")},
			{Key: aws.String("ExecutionStage"), Value: aws.String("Executing")},
			//{Key: aws.String("Status"), Value: aws.String("InProgress")},
		},
		InstanceId: aws.String(vmid),
	})
	if err != nil {
		return true
	}

	if len(cmdInvocation.Commands) > 0 {
		return true
	}

	return false
}
func SSMRunCommand(shellCmds []string) string {
	ssmSes := ssm.New(sess)
	vmid, err := getVmID()
	if err != nil {
		return err.Error()
	}
	if curentCmdInProgress(ssmSes, vmid) {
		return "Please try again latest. CmdId still in progress"
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
	for {
		switch {
		case err != nil:
			return fmt.Sprintf(
				"GetCommandInvocation err:\nCommandId: %s; InstanceId: %v;\n\n%s",
				*sendCmd.Command.CommandId, vmid, err.Error())
		case cmdInvocation != nil:
			switch status := *cmdInvocation.Status; status {
			case ssm.CommandInvocationStatusCancelling, ssm.CommandInvocationStatusInProgress, ssm.CommandInvocationStatusPending:
				println("CommandInvocationStatus " + *sendCmd.Command.CommandId + " : " + status)
				time.Sleep(500 * time.Millisecond)
			default:
				println("CommandInvocationStatus: " + status)
				return fmt.Sprintf("Executed:\n```\n%v\n```\nResult:\n%s", strings.Join(shellCmds[:], "\n"), *cmdInvocation.StandardErrorContent+*cmdInvocation.StandardOutputContent)
			}
		default:
			println("Command Invocation is nil")
			time.Sleep(500 * time.Millisecond)
		}

	}
	//return "unexpected behavior"

}

func tgShowVerVM(b *tb.Bot, m *tb.Message) {
	if !ifStackCreated() {
		b.Reply(m, "VM stack is not exist")
		return
	}
	cmd := []string{
		"PATH=/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin:/root/bin",
		"kf5-config --version 2>/dev/null",
		"uname -a && date -u",
		"pkg info -a | grep kde5",
		"echo -n Private ip:",
		"curl -s 169.254.169.254/latest/meta-data/local-ipv4",
		"echo",
		"echo -n Public ip:",
		"curl -s 169.254.169.254/latest/meta-data/public-ipv4",
	}
	ssmStdOut := SSMRunCommand(cmd)

	b.Reply(m, ssmStdOut, &tb.SendOptions{DisableWebPagePreview: true})

}

func tgKdePatchVM(b *tb.Bot, m *tb.Message) {
	if !ifStackCreated() {
		b.Reply(m, "VM stack is not exist")
		return
	}
	cmd := []string{
		"PATH=/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin:/root/bin",
		"mkdir -p /usr/local/etc/pkg/repos",
		"cp /etc/pkg/FreeBSD.conf /usr/local/etc/pkg/repos/",
		"sed -I bak s/quarterly/latest/ /usr/local/etc/pkg/repos/FreeBSD.conf",
		"pkg update",
		"pkg upgrade -y",
		"echo patched",
		"reboot",
	}
	ssmStdOut := SSMRunCommand(cmd)

	b.Reply(m, fmt.Sprintf("%s\n\nPlease run %s again", ssmStdOut, TbCmdOnVmShowVer), &tb.SendOptions{DisableWebPagePreview: true})
}

func tgMonitorVM(b *tb.Bot, m *tb.Message) {
	if !ifStackCreated() {
		b.Reply(m, "VM stack is not exist")
		return
	}
	cmd := []string{
		"PATH=/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin:/root/bin",
		"top -b -z -w -n 10",
	}
	ssmStdOut := SSMRunCommand(cmd)
	b.Reply(m, ssmStdOut, &tb.SendOptions{DisableWebPagePreview: true})
}
