package main

type Conversation struct{
	ID int
	Other_ID int
	Participant_IDs []int
	Created_at string
}

type Message struct{
	ID int 
	Conversation_ID int 
	Sender_ID int 
	Text string
	Created_at string

}

type Participant struct{
	Conversation_ID int 
	User_ID int
	Role Role
}

type Role struct{
	Passenger bool
	Driver bool
	Dispatcher bool
	Admin bool
}

