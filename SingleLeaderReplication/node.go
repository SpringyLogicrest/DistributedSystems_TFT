package main

import (
	pb "SingleLeaderReplication/proto"
	"sync"
	"time"
)

const rpcTimeout = 500 * time.Millisecond

type Node struct {
	pb.UnimplementedAuctionServiceServer

	ID    int
	Peers map[int]string
	mu    sync.Mutex



	isLeader      bool
	leaderID      int
	highestBid    int64
	highestBidder int64
	bidders 	 map[int64]bool
	auctionEnded  bool
	endTime      time.Time

	alivemu sync.Mutex
	alive map[int64]bool
	

}

func NewNode(id int, leaderID int) *Node {
	return &Node{
		ID:       id,
		isLeader: id == leaderID,
		leaderID: leaderID,
	}
}
