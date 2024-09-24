package grpcProvider

import (
	context "context"
	"encoding/json"
	"fmt"
	"gRPCClient/models"
	"gRPCClient/utils"
	"log"
	"runtime"
	"time"

	"github.com/google/uuid"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	SendChannel chan models.Message
	SC          ServicesClient
	Stream      Services_ConnectClient
	Context     context.Context
	Timer       time.Timer
}

func InitializeConnection() *GRPCClient {

	conn, err := grpc.Dial("127.0.0.1:50000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		utils.LogFatal("InitializeConnection", "failed to connect to gRPC server at localhost:50051", nil, err)
	}

	c := NewServicesClient(conn)

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)

	defer cancelFunc()

	var clientContext models.ClientContext

	clientContext.Platform = runtime.GOOS
	clientContext.ClientID = uuid.NewString()

	// add any name
	clientContext.Name = "the-shiva-dev"

	// Fetch the computer system details.
	computerSystem, err := utils.FetchComputerSystemDetails()
	if err != nil {
		utils.LogError("InitializeConnection", "error fetching computer system details", "", err)
	}

	clientContext.ComputerSystem = computerSystem

	// Marshal the struct into JSON format
	clientContextBytes, err := json.Marshal(clientContext)
	if err != nil {
		utils.LogError("InitializeConnection", "error converting client details to byte", fmt.Sprintf("client Details: %v", clientContext), err)
	}

	stream, err := c.Connect(context.Background())
	if err != nil {
		utils.LogError("InitializeConnection", "error connecting to tcp server ", "", err)
	}

	err = stream.Send(&Message{Message: clientContextBytes, MessageType: "register"})
	if err != nil {
		utils.LogError("InitializeConnection", "error sending message to server ", "", err)
	}

	utils.LogInfo("InitializeConnection", "sending client system information to server for a stream connection request", "", err)

	return &GRPCClient{
		SC:          c,
		Context:     ctx,
		Stream:      stream,
		Timer:       time.Timer{},
		SendChannel: make(chan models.Message, 1),
	}
}

func (gc *GRPCClient) InitChat() error {

	go gc.Write()
	go gc.InitialPing()
	gc.Read()
	return nil
}

func (gc *GRPCClient) Write() {
	count := 0
	// Indefinitely listen on the send channel for any messages that have to be sent to the server.
	for {
		count++

		fmt.Println("iteration ", count)
		select {

		case sendMessage, ok := <-gc.SendChannel:

			// On calling the Close() function, the SendChannel is closed which will return a false value to `ok`.
			// Therefore the loop will break from the else case and the function will return, closing the function.
			if !ok {
				return
			}

			// Send the message in the stream.
			err := gc.Stream.Send(&Message{
				MessageType: sendMessage.MessageType,
				Message:     sendMessage.Message,
			})
			if err != nil {
				utils.LogError("Write", "error sending message in the message stream", fmt.Sprintf("Message: %v", sendMessage), err)
				return
			}
			log.Print("sent message successfully ", fmt.Sprintf("Message Type: %s", sendMessage.MessageType))
			utils.LogInfo("Write", "sent the message successfully", fmt.Sprintf("Message Type: %s", sendMessage.MessageType), nil)

		// Handle the timeout.
		case <-gc.Timer.C:

			// Close the stream.
			utils.LogInfo("Write", "timeout reached, attempting to send the CloseSend message to close the message stream successfully", "", nil)
			return
		}

	}
}

func (gc *GRPCClient) InitialPing() {
	// Send the message in the stream.
	utils.LogInfo("InitialPing", fmt.Sprintf("Sending Initial message : %s\n", models.PingMessage), "", nil)
	err := gc.Stream.Send(&Message{
		MessageType: models.PingMessageType,
		Message:     []byte(models.PingMessage),
	})
	if err != nil {
		utils.LogError("InitialPing", "error sending ping message in the message stream", fmt.Sprintf("Message: %v", models.PingMessage), err)
		return
	}

}

// Stream reader reads the messages received over the gRPC connection in connection stream.
func (gc *GRPCClient) Read() error {

	for {

		// Receive the message from the stream.
		message, err := gc.Stream.Recv()
		if err != nil {
			utils.LogError("Read", "error reading message from stream", "", err)
			return err
		}

		utils.LogInfo("Read", "received message", "", fmt.Sprintf("Message: %v, MessageType: %v", string(message.Message), message.MessageType))

		// comment the log below if dont want to see the messages on console
		log.Print("received message successfully ", fmt.Sprintf("Message: %v, MessageType: %v", string(message.Message), message.MessageType))

		go gc.ProcessServerMessage(message.MessageType, message.Message)
	}
}

func (nc *GRPCClient) ProcessServerMessage(messageType string, message []byte) {

	switch messageType {
	case models.PongMessageType:
		go nc.ProcessPong()
	}

}

func (nc *GRPCClient) ProcessPong() {
	time.Sleep(5 * time.Second)
	utils.LogInfo("ProcessPong", fmt.Sprintf("Sending  message : %s\n", models.PingMessage), "", nil)
	nc.Timer = time.Timer{C: time.After(10 * time.Second)}
	nc.SendChannel <- models.Message{
		Message:     []byte(models.PingMessage),
		MessageType: models.PingMessageType,
	}
}
