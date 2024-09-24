package main

import (
	"gRPCClient/providers/grpcProvider"
)

func main() {
	// initializing connection to the gRPC server.
	GRPCClient := grpcProvider.InitializeConnection()

	// initializing connection to the stream of gRPC server.
	GRPCClient.InitChat()
}
