package models

const (
	// Health Check
	PingMessage     string = "Ping"
	PingMessageType string = "Ping Message Type"
	PongMessage     string = "Pong"
	PongMessageType string = "Pong Message Type"

	// Registration
	RegistrationRequestMessageType    string = "Registration Request Message Type"
	RegistrationFailedMessageType     string = "Registration Failed Message Type"
	RegistrationSuccessfulMessageType string = "Registration Successful Message Type"
)

type ClientContext struct {
	Platform       string `json:"platform" bson:"platform"`
	ComputerSystem `bson:",inline"`
	ClientID       string `json:"clientID" bson:"clientID"`
	Name           string `json:"name" bson:"name"`
}

// Details of the computer system.
type ComputerSystem struct {
	CurrentLoggedInUser string `json:"currentLoggedInUser" bson:"currentLoggedInUser"`
	Domain              string `json:"domain" bson:"domain"`
	Hostname            string `json:"hostname" bson:"hostname"`
}

// Message strcuture to facilitate communication.
type Message struct {
	Message     []byte `json:"message"`
	MessageType string `json:"messageType"`
}
