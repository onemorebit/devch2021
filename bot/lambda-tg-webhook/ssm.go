package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"strings"
	"time"
)

func SSMRunCommand(shellCmds []string) string {
	ssmSes := ssm.New(sess)
	vmid, err := GetVmID()
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
	time.Sleep(time.Second)
	cmdInvocation, err := ssmSes.GetCommandInvocation(&ssm.GetCommandInvocationInput{
		CommandId:  sendCmd.Command.CommandId,
		InstanceId: aws.String(vmid),
	})
	if err != nil {
		return fmt.Sprintf(
			"GetCommandInvocation err:\n"+
				"CommandId: %s; InstanceId: %s;\n\n%s",
			*sendCmd.Command.CommandId, vmid, err.Error())
	}
	//return "stderr:\n" + *cmdInvocation.StandardErrorContent + "\nstdout:\n" + *cmdInvocation.StandardOutputContent
	return fmt.Sprintf("Executed:\n```\n%v\n```\nResult:\n%s", strings.Join(shellCmds[:], "\n"), *cmdInvocation.StandardErrorContent+*cmdInvocation.StandardOutputContent)

}
