package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
	

	pb "ricart-agrawala/proto"

	"google.golang.org/grpc"
)

// Hardcoded addresses for the three nodes.
var peers = map[int64]string{
	1: "localhost:50051",
	2: "localhost:50052",
	3: "localhost:50053",
}

// Starts the node and handles user commands.
func main() {
	id := flag.Int64("id", 1, "Node ID")
	flag.Parse()

	// Find the address for this node.
	address, exists := peers[*id]
	if !exists {
		log.Fatalf("Unknown node ID: %d", *id)
	}

	// this is to create a new map of peers excluding the current node
	others := make(map[int64]string)
	for peerID, peerAddress := range peers {
		if peerID != *id {
			others[peerID] = peerAddress
		}
	}
	node := &Node{
		ID: *id,
		Peers: others,
	}

	// Start listening for connections.
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create and register the gRPC server.
	grpcServer := grpc.NewServer()
	pb.RegisterMutexServiceServer(
		grpcServer,
		NewMutexServer(node),
	)

	// Run the server in the background.
	go func() {
		log.Printf(
			"Node %d: listening on %s",
			node.ID,
			address,
		)

		if err := grpcServer.Serve(listener); err != nil {
			log.Printf("gRPC server stopped: %v", err)
		}
	}()

	fmt.Println("Commands: probe, quit")

	scanner := bufio.NewScanner(os.Stdin)

	// Read commands from the terminal.
	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			break
		}

		command := strings.TrimSpace(scanner.Text())

		switch command {
		case "probe":
			RequestCriticalSectionAccess(node)
			time.Sleep(10 * time.Second) // this is the time inside the cs
			ExitCriticalSection(node)
			

		case "quit":
			// Stop the server and exit.
			grpcServer.GracefulStop()
			return

		default:
			fmt.Println("Unknown command. Use: probe, quit")
		}
	}
	// Check if an error occurred while reading input.
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v", err)
	}

	grpcServer.GracefulStop()
}
