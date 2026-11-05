//recieves and handles grpc messages

package main

import (
	"context"
	"log"

	pb "ricart-agrawala/proto"
)

// Handles requests from other nodes.
type MutexServer struct {
	pb.UnimplementedMutexServiceServer
	node *Node
}

// Creates a server associated with a node.
func NewMutexServer(node *Node) *MutexServer {
	return &MutexServer{
		node: node,
	}
}

// Receives a request and sends back a reply.
func (s *MutexServer) RequestAccess(
	ctx context.Context,
	req *pb.Request,
) (*pb.Reply, error) {

	clock := s.node.ReceiveTimestamp(req.Timestamp)

	log.Printf(
		"Node %d: received REQUEST from Node %d (timestamp=%d, clock=%d)",
		s.node.ID,
		req.NodeId,
		req.Timestamp,
		clock,
	)

	// Temp response for testing gRPC communication (Ricart-Agrawala permission logic not implemented yet).
	log.Printf(
		"Node %d: sending provisional REPLY to Node %d",
		s.node.ID,
		req.NodeId,
	)

	return &pb.Reply{Granted: true}, nil
}
