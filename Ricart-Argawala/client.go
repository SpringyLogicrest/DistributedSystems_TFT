package main

import (
	"context"
	"fmt"
	"log"

	pb "ricart-agrawala/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Sends a test request to another node.
func SendRequest(
	node *Node,
	peerID int64,
	peerAddress string,
	timestamp int64,

) error {

	// Connect to the other node.
	conn, err := grpc.NewClient(
		peerAddress,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)

	if err != nil {
		return fmt.Errorf("create gRPC client: %w", err)
	}
	defer conn.Close()

	client := pb.NewMutexServiceClient(conn)

	ctx := context.Background()

	log.Printf(
		"Node %d: sending REQUEST to Node %d (timestamp=%d)",
		node.ID,
		peerID,
		timestamp,
	)

	// Send the request using gRPC.
	reply, err := client.RequestAccess(
		ctx,
		&pb.Request{
			NodeId:    node.ID,
			Timestamp: timestamp,
		},
	)

	if err != nil {
		return fmt.Errorf(
			"request to node %d failed: %w",
			peerID,
			err,
		)
	}

	// Advance the local clock after receiving the reply. Currently, the reply does not contain a timestamp.
	node.NextTimestamp()

	log.Printf(
		"Node %d: received REPLY from Node %d (granted=%t)",
		node.ID,
		peerID,
		reply.Granted,
	)

	return nil
}
