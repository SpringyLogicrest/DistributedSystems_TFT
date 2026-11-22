//here we implement methods such as Bid() and Result()

package main

import (
	pb "SingleLeaderReplication/proto"
	"context"
)

type AuctionServer struct {
	pb.UnimplementedAuctionServiceServer
	node *Node
}

func (s *AuctionServer) Bid(ctx context.Context, req *pb.BidRequest) (*pb.BidResponse, error) {

	s.node.mu.Lock()

	if !s.node.isLeader {
		s.node.mu.Unlock()
		return &pb.BidResponse{Ack: "exception"}, nil
	}

	if s.node.auctionEnded {
		s.node.mu.Unlock()
		return &pb.BidResponse{Ack: "fail"}, nil
	}

	if req.Amount <= s.node.highestBid {
		s.node.mu.Unlock()
		return &pb.BidResponse{Ack: "fail"}, nil
	}

	s.node.highestBid = req.Amount
	s.node.highestBidder = req.BidderId
	s.node.bidders[req.BidderId] = true

	s.node.mu.Unlock()

	//send updated auction state to the backup nodes
	s.node.SendUpdate()

	return &pb.BidResponse{Ack: "success"}, nil
}

func (s *AuctionServer) Result(
	ctx context.Context,
	req *pb.ResultRequest,
) (*pb.ResultResponse, error) {

	s.node.mu.Lock()
	defer s.node.mu.Unlock()

	return &pb.ResultResponse{
		HighestBid:    s.node.highestBid,
		HighestBidder: s.node.highestBidder,
		AuctionEnded:  s.node.auctionEnded,
	}, nil
}
