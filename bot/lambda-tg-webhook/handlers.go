package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	tb "gopkg.in/tucnak/telebot.v2"
)

var sess = GetSharedConfigSession()
var stackName = "freebsd-devch2021"

func tgCreateVM(b *tb.Bot, m *tb.Message) {
	if !ifStackCreated() {
		b.Send(m.Chat, "Creating VM")
		createVM()
		return
	}
	b.Send(m.Chat, "VM stack is already exist")

	id, err := GetVmID()
	if err != nil {
		println("getvmid err: " + err.Error())
		return
	}
	if id != "" {
		b.Send(m.Chat, "EC2 id: "+id)
		return
	}
	b.Send(m.Chat, "Please Wait. Creating VM in progress")
}
func tgDestroyVM(b *tb.Bot, m *tb.Message) {
	if !ifStackCreated() {
		b.Send(m.Chat, "VM stack is not exist")
		return
	}
	scv := cloudformation.New(sess)
	_, err := scv.DeleteStack(
		&cloudformation.DeleteStackInput{
			StackName: aws.String(stackName)},
	)
	if err != nil {
		b.Send(m.Chat, "CFN stack can not be deleted: "+err.Error())
		return
	}
	b.Send(m.Chat, "Stack will be deleted soon")

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

func GetSharedConfigSession() *session.Session {
	return session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
}

func createVM() {
	scv := cloudformation.New(sess)
	input := &cloudformation.CreateStackInput{
		StackName: aws.String(stackName),
		OnFailure: aws.String("ROLLBACK"),
		Capabilities: []*string{
			aws.String("CAPABILITY_IAM"),
			aws.String("CAPABILITY_NAMED_IAM"),
		},
		TemplateBody:     aws.String(string(freebsdCform)),
		TimeoutInMinutes: aws.Int64(20),
	}
	resp, err := scv.CreateStack(input)
	if err != nil {
		println(err)
	}

	println(awsutil.Prettify(resp))
}

func GetStacks() (*cloudformation.DescribeStacksOutput, error) {
	scv := cloudformation.New(sess)
	return scv.DescribeStacks(&cloudformation.DescribeStacksInput{StackName: aws.String(stackName)})
}

func ifStackCreated() bool {
	res, err := GetStacks()
	if err != nil {
		return false
	}
	if res == nil {
		return false
	}
	if len(res.Stacks) == 1 {
		return true
	}
	return false
}

func GetVmID() (instanceid string, err error) {

	res, err := GetStacks()
	if err != nil {
		return
	}
	outputs := map[string]string{}
	for _, stack := range res.Stacks {
		for _, output := range stack.Outputs {
			outputs[*output.OutputKey] = *output.OutputValue
		}
	}
	instanceid, ok := outputs["InstanceId"]
	if ok {
		return
	}
	return
}
