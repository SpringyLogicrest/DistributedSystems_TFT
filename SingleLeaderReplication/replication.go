//this file makes it possible to send updates/information to backup nodes

package main

import (
	pb "SingleLeaderReplication/proto"
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ReplicationServer struct {
	pb.UnimplementedReplicationServiceServer
	node *Node
}


// this function is used to send the state of the auction to the backup nodes
// it is called by the leader when it receives a new bid
func dial(address string) (pb.ReplicationServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	return pb.NewReplicationServiceClient(conn), conn, nil
}

// this is the leader side
func (n *Node) SendUpdate() {

	n.mu.Lock()

	bidderIDs := make([]int64, 0, len(n.bidders))

	for bidderID, active := range n.bidders {
		if active {
			bidderIDs = append(bidderIDs, bidderID)
		}
	}

	state := &pb.State{
		HighestBid:    n.highestBid,
		HighestBidder: n.highestBidder,
		BidderIds:     bidderIDs,
	}

	n.mu.Unlock()

	for peerID, address := range n.Peers {
		//create a gRPC client connection
		conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			fmt.Println("Failed to create client for node", peerID, err)
			continue
		}

		client := pb.NewReplicationServiceClient(conn)

		//set a timeout for the RPC
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		//send the state to the follower
		_, err = client.Replicate(ctx, state)
		cancel()
		conn.Close()

		if err != nil {
			fmt.Println("Failed to replicate to node", peerID, err)
			continue
		}

		fmt.Println("State replicated to node", peerID)
	}
}

func (s *ReplicationServer) Replicate(
	ctx context.Context,
	req *pb.State,
) (*pb.Empty, error) {
	s.node.mu.Lock()
	defer s.node.mu.Unlock()

	s.node.highestBid = req.HighestBid
	s.node.highestBidder = req.HighestBidder

	fmt.Printf(
		"Node %d replicated state: Highest bid = %d, Highest bidder = %d\n",
		s.node.ID, s.node.highestBid, s.node.highestBidder,
	)

	s.node.bidders = make(map[int64]bool)

	for _, bidderID := range req.BidderIds {
		s.node.bidders[bidderID] = true
	}

	return &pb.Empty{}, nil
}

