package grpcProvider

import (
	"gRPCClient/models"
	"gRPCClient/utils"
	context "context"
	"encoding/json"
	"fmt"
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

// func InitializeConnection() *GRPCClient {

// 	// grpc.

// 	conn, err := grpc.Dial("127.0.0.1:50000", grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil {
// 		log.Fatalf("failed to connect to gRPC server at localhost:50051: %v", err)
// 	}
// 	// defer conn.Close()
// 	c := NewServicesClient(conn)

// 	ctx, _ := context.WithTimeout(context.Background(), time.Second)

// 	// c.Connect(ctx)
// 	// defer cancel()
// 	var clientContext models.ClientContext
// 	// var err error

// 	fmt.Println("111111111111")

// 	connectionContext, _ := context.WithTimeout(context.Background(), time.Second)

// 	clientContext.Platform = runtime.GOOS
// 	clientContext.ClientID = uuid.NewString()
// 	clientContext.Name = "shiv"

// 	// Fetch the computer system details.
// 	computerSystem, err := utils.FetchComputerSystemDetails()
// 	if err != nil {
// 		log.Println("SendClientContext", "error fetching computer system details", "", err)
// 		// return nil
// 	}

// 	clientContext.ComputerSystem = computerSystem

// 	// Marshal the struct into JSON format
// 	clientContextBytes, err := json.Marshal(clientContext)
// 	if err != nil {
// 		log.Println("SendClientContext", "error converting client details to byte array", fmt.Sprintf("client Details: %v", clientContext), err)
// 		// return err
// 	}
// 	fmt.Println("22222222222222")
// 	stream, err := c.Connect(connectionContext)
// 	if err != nil {
// 		log.Println("SendClientContext", "error connecting to tcp server ", "", err)
// 		// return err
// 	}

// 	// Send the data to the server to complete the registration process.
// 	err = stream.Send(&Message{
// 		Message:     clientContextBytes,
// 		MessageType: string(models.RegistrationRequestMessageType),
// 	})

// 	fmt.Println("3333333333")

// 	if err != nil {
// 		log.Println("SendClientContext", "error sending data in the bi-directional stream established with the server", fmt.Sprintf("client Context: %v", clientContext), err)
// 		// return err
// 	}

// 	// gc.Stream = stream

// 	return &GRPCClient{
// 		SC:      c,
// 		Context: ctx,
// 		Stream:  stream,
// 	}
// }

func InitializeConnection() *GRPCClient {

	fmt.Println("111111111111111")

	conn, err := grpc.Dial("127.0.0.1:50000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to gRPC server at localhost:50051: %v", err)
	}
	// defer conn.Close()
	c := NewServicesClient(conn)

	ctx, _ := context.WithTimeout(context.Background(), time.Second)

	var clientContext models.ClientContext
	// var err error

	fmt.Println("111111111111")

	clientContext.Platform = runtime.GOOS
	clientContext.ClientID = uuid.NewString()
	clientContext.Name = "shiv"

	// Fetch the computer system details.
	computerSystem, err := utils.FetchComputerSystemDetails()
	if err != nil {
		log.Println("SendClientContext", "error fetching computer system details", "", err)
		// return err
	}

	clientContext.ComputerSystem = computerSystem

	// Marshal the struct into JSON format
	clientContextBytes, err := json.Marshal(clientContext)
	if err != nil {
		log.Println("SendClientContext", "error converting client details to byte array", fmt.Sprintf("client Details: %v", clientContext), err)
		// return err
	}

	stream, err := c.Connect(context.Background())
	if err != nil {
		log.Println("SendClientContext", "error connecting to tcp server ", "", err)
		// return err
	}

	fmt.Println("2222222222")

	time.Sleep(3 * time.Second)

	err = stream.Send(&Message{Message: clientContextBytes, MessageType: "csdc"})
	if err != nil {
		log.Println("SendClientContext", "error sending message from server ", "", err)
		// return err
	}

	return &GRPCClient{
		SC:          c,
		Context:     ctx,
		Stream:      stream,
		Timer:       time.Timer{},
		SendChannel: make(chan models.Message, 1),
	}
}

func (gc *GRPCClient) SendClientContext() error {

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
				log.Println("Write", "error sending message in the message stream", fmt.Sprintf("Message: %v", sendMessage), err)
				return
			}
			// log.Println("Write", "sent the message successfully", sendMessage.MessageType, sendMessage)
			log.Println("Write", "sent the message successfully", fmt.Sprintf("Message Type: %s", sendMessage.MessageType), nil)

		// Handle the timeout.
		case <-gc.Timer.C:

			// Close the stream.
			log.Println("Write", "timeout reached, attempting to send the CloseSend message to close the message stream successfully", "", nil)
			return
		}

	}
}

func (gc *GRPCClient) InitialPing() {
	// Send the message in the stream.
	log.Printf("Sending Initial message : %s\n", models.PingMessage)
	err := gc.Stream.Send(&Message{
		MessageType: models.PingMessageType,
		Message:     []byte(models.PingMessage),
	})
	if err != nil {
		log.Println("Write", "error sending ping message in the message stream", fmt.Sprintf("Message: %v", models.PingMessage), err)
		return
	}

}

// Stream reader reads the messages received over the gRPC connection in connection stream.
func (gc *GRPCClient) Read() error {

	for {

		// Receive the message from the stream.
		message, err := gc.Stream.Recv()
		if err != nil {
			log.Println("Read error reading message from stream", "", nil)
			return err
		}

		// var clientContext models.ClientContext

		// err = json.Unmarshal(message.Message, &clientContext)
		// if err != nil {
		// 	log.Println("Read error decoding message from stream", "", err.Error())
		// 	return err
		// }

		log.Println("Read received message", fmt.Sprintf("Message: %v", string(message.Message)), fmt.Sprintf("MessageType: %v", message))

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
	log.Printf("Sending  message : %s\n", models.PingMessage)
	fmt.Println("666666666666666")
	nc.Timer = time.Timer{C: time.After(10 * time.Second)}
	fmt.Println("77777777777777777")

	nc.SendChannel <- models.Message{
		Message:     []byte(models.PingMessage),
		MessageType: models.PingMessageType,
	}
}

// func (gc *GRPCClient) terminalWriter() {

// 	for {
// 		var message models.Message
// 		var text string
// 		fmt.Scan(&text)
// 		if text == "exit" {
// 			break
// 		}
// 		message.Message = []byte(text)
// 		message.MessageType = "shiv"
// 		gc.SendChannel <- message
// 	}

// }
