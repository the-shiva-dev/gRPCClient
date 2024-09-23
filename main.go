package main

import (
	"fmt"

	"gRPCClient/providers/grpcProvider"
)

func main() {
	fmt.Println("hello gRPC")
	GRPCClient := grpcProvider.InitializeConnection()

	GRPCClient.SendClientContext()
}
