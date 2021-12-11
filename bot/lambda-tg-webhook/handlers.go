package main

import (
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
func GetVmID() (instanceid string, err error) {
	scv := cloudformation.New(sess)
	res, err := scv.DescribeStacks(&cloudformation.DescribeStacksInput{StackName: aws.String(stackName)})
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
