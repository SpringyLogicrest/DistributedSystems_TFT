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

	s.node.Mu.Lock()
	shouldDefer := false

	if s.node.InCS {
		shouldDefer = true
	} else if s.node.Requesting {
		shouldDefer = HasPriority(
			s.node.RequestTS,
			s.node.ID,
			req.Timestamp,
			req.NodeId)
	}

	//a channel for a deferred request
	var waitCh chan struct{}

	if shouldDefer {
		waitCh = make(chan struct{})

		s.node.Deferred = append(
			s.node.Deferred,
			waitCh,
		)
	}
	s.node.Mu.Unlock()

	if shouldDefer {
		log.Printf("Node %d: deffering reply to Node %d", s.node.ID, req.NodeId)

		select {
		case <-waitCh:
			//Release function has signaled permission
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	log.Printf("Node %d: sending REPLY to Node %d", s.node.ID, req.NodeId)
	return &pb.Reply{Granted: true}, nil
}
