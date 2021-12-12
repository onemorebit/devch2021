package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

var sess = getSharedConfigSession()
var stackName = "freebsd-devch2021"
var curentStack *cloudformation.Stack

func getSharedConfigSession() *session.Session {
	return session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
}

func getStacks() (*cloudformation.DescribeStacksOutput, error) {
	scv := cloudformation.New(sess)
	return scv.DescribeStacks(&cloudformation.DescribeStacksInput{StackName: aws.String(stackName)})
}

func ifStackCreated() bool {
	res, err := getStacks()
	if err != nil {
		return false
	}
	if res == nil {
		return false
	}
	if len(res.Stacks) == 1 {
		curentStack = res.Stacks[0]
		return true
	}
	return false
}

func getVmID() (instanceid string, err error) {

	res, err := getStacks()
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
	if instanceid == "" {
		stackStatus := "undefined"
		if curentStack != nil {
			stackStatus = *curentStack.StackStatus
		}
		return "", fmt.Errorf("Empty instance id, stack status: " + stackStatus)
	}
	return
}
