package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"gopkg.in/tucnak/telebot.v2"
)

// i will be happy to use SQS FIFO for the Q and lock features,
// but have not time to implement this things
// so, we will use the global variable
// I Understand the Risks

type GlobalFreeBSDState int

const (
	Absent GlobalFreeBSDState = iota
	Creating
	Created
	Destroying
)

var CurentFreeBSDState GlobalFreeBSDState
var sess = GetSharedConfigSession()
var stackName = "freebsd-devch2021"

func (d GlobalFreeBSDState) String() string {
	return [...]string{"Absent", "Creating", "Created", "Destroying"}[d]
}

func tgCreateVM(b *telebot.Bot, m *telebot.Message) {

	switch CurentFreeBSDState {
	case Absent:
		b.Send(m.Chat, "Creating VM")
		CurentFreeBSDState = Creating
		createVM()
	default:
		id, err := GetVmID()
		if err != nil {
			println("getvmid err: " + err.Error())
			return
		}
		if id != "" {
			CurentFreeBSDState = Created
			b.Send(m.Chat, "EC2 id: "+id+"\nCurent state: "+CurentFreeBSDState.String())
			return
		}
		b.Send(m.Chat, "Please Wait. Curent state: "+CurentFreeBSDState.String())

	}
}
func tgDestroyVM(b *telebot.Bot, m *telebot.Message) {
	if !ifStackCreated() {
		b.Send(m.Chat, "VM stack is not exist")
	}
	scv := cloudformation.New(sess)
	delOut, err := scv.DeleteStack(
		&cloudformation.DeleteStackInput{
			StackName: aws.String(stackName)},
	)
	if err != nil {
		b.Send(m.Chat, "Delete stack issues: "+err.Error())
		return
	}
	b.Send(m.Chat, "Delete stack success: "+delOut.String())

}

func tgShowVerVM(b *telebot.Bot, m *telebot.Message) {
	if !ifStackCreated() {
		b.Send(m.Chat, "VM stack is not exist")
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

func tgKdePatchVM(b *telebot.Bot, m *telebot.Message) {
	if !ifStackCreated() {
		b.Send(m.Chat, "VM stack is not exist")
	}
	cmd := []string{
		"PATH=/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin:/root/bin",
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
func tgMonitorVM(b *telebot.Bot, m *telebot.Message) {
	if !ifStackCreated() {
		b.Send(m.Chat, "VM stack is not exist")
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
