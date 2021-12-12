package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	tb "gopkg.in/tucnak/telebot.v2"
)

func tgCreateVM(b *tb.Bot, m *tb.Message) {
	if !ifStackCreated() {
		b.Send(m.Chat, "Creating VM")
		createVM()
		return
	}
	b.Send(m.Chat, "VM stack is already exist")

	id, err := getVmID()
	if err != nil {
		println("getvmid err: " + err.Error())
		return
	}
	if id != "" {
		b.Send(m.Chat, "EC2 id: "+id)
		return
	}
	b.Send(m.Chat, "Please Wait. Creating VM in progress. Stack status: "+*curentStack.StackStatus)
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
	b.Send(m.Chat, "Stack will be deleted soon: "+*curentStack.StackStatus)

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
