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
	defer s.node.mu.Unlock()

	if !s.node.IsLeader {
		return &pb.BidResponse{Ack: "exception"}, nil
	}

	if s.node.AuctionEnded {
		return &pb.BidResponse{Ack: "fail"}, nil
	}

	if req.Amount <= s.node.HighestBid {
		return &pb.BidResponse{Ack: "fail"}, nil
	}

	s.node.HighestBid = req.Amount
	s.node.HighestBidder = req.BidderId

	return &pb.BidResponse{Ack: "success"}, nil
}

func (s *AuctionServer) Result(
	ctx context.Context,
	req *pb.ResultRequest,
) (*pb.ResultResponse, error) {

	s.node.mu.Lock()
	defer s.node.mu.Unlock()

	return &pb.ResultResponse{
		HighestBid:    s.node.HighestBid,
		HighestBidder: s.node.HighestBidder,
		AuctionEnded:  s.node.AuctionEnded,
	}, nil
}
