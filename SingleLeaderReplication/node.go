// To the TA:
// we are trying to make a bully election where the lowest ID  wins 

package main

import (
	pb "SingleLeaderReplication/proto"
	"sync"
	"time"
)

const rpcTimeout = 500 * time.Millisecond
const heartbeatInterval = 200 * time.Millisecond
const leaderTimeout = 1 * time.Second
const auctionDuration = 100 * time.Second
const startupDelay = 5 * time.Second




type Node struct {
	pb.UnimplementedAuctionServiceServer

	ID    int
	Peers map[int]string
	mu    sync.Mutex
	bidMu sync.Mutex

	isLeader      bool
	leaderID      int
	highestBid    int64
	highestBidder int64
	bidders       map[int64]bool
	electing      bool
	lastHeartbeat time.Time
	auctionEnded  bool
	endTime       time.Time

}

func NewNode(id int, leaderID int) *Node {
	return &Node{
		ID:       id,
		isLeader: id == leaderID,
		leaderID: leaderID,

		Peers:   make(map[int]string),
		bidders: make(map[int64]bool),
	}
}

// this function bascially checks if the auction has ended or not
func (n *Node) ended() bool {
	return !n.endTime.IsZero() && time.Now().After(n.endTime)
}


// function is for the leader to send the state of the auction to the followers
func (n *Node) snapshot() *pb.State {
	ids := make([]int64, 0, len(n.bidders))
	for id := range n.bidders {
		ids = append(ids, id)
	}
	return &pb.State{
		HighestBid:    n.highestBid,
		HighestBidder: n.highestBidder,
		BidderIds:     ids,
		EndTimeUnixMs: n.endTime.UnixMilli(),
		LeaderId:      int64(n.leaderID),
	}


}


// so this function is used  to apply the state of the auction to the node when it receives a replication from the leader
func (n *Node) applyState(state *pb.State) {
	n.highestBid = state.HighestBid
	n.highestBidder = state.HighestBidder
	if state.EndTimeUnixMs != 0 {
		n.endTime = time.UnixMilli(state.EndTimeUnixMs)
	}
	n.bidders = make(map[int64]bool)
	for _, id := range state.BidderIds {
		n.bidders[id] = true
	}
}

