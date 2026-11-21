package main

import "sync"

type Node struct {
	ID            int
	IsLeader      bool
	LeaderID      int
	HighestBid    int64
	HighestBidder int64
	AuctionEnded  bool

	mu sync.Mutex
}

func NewNode(id int, leaderID int) *Node {
	return &Node{
		ID:       id,
		IsLeader: id == leaderID,
		LeaderID: leaderID,
	}
}
