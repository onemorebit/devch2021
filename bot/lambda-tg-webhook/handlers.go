package main

import "gopkg.in/tucnak/telebot.v2"

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

func (d GlobalFreeBSDState) String() string {
	return [...]string{"Absent", "Creating", "Created", "Destroying"}[d]
}

func tgCreateVM(b *telebot.Bot, m *telebot.Message) {

	// Read lock
	//
	b.Send(m.Chat, "CurentFreeBSDState: "+CurentFreeBSDState.String())
	if CurentFreeBSDState < Destroying {
		CurentFreeBSDState += 1
	} else {
		CurentFreeBSDState = Absent
	}
}
