//here is the logic for choosing a new leader
// our goal is to make a bully election where the lowest ID wins

package main

import (
	pb "SingleLeaderReplication/proto"
	"context"
	"fmt"
	"time"
)

// we start by making failure detection

// this function is used to run the failure detection in a separate goroutine
// it runsfor every node and checks if the leader is alive or not
// and it runs forever until the program is terminated

func (n *Node) RunFailureDetection() {
	for {
		time.Sleep(heartbeatInterval)

		n.mu.Lock()
		isLeader := n.isLeader
		silentFor := time.Since(n.lastHeartbeat)
		n.mu.Unlock()

		if isLeader {
			n.sendHeartbeats()
		} else if silentFor > leaderTimeout {
			fmt.Println("Node", n.ID, ": leader", n.leaderID, "is dead, starting election")
			n.startElection()
		}
	}
}

// the leader sends heartbeats to the followers to let them know that it is still alive
// its basically a message saying "im alive " 
func (n *Node) sendHeartbeats() {
	n.mu.Lock()
	state := n.snapshot()
	n.mu.Unlock()

	for _, address := range n.Peers {
		client, conn, err := dial(address)
		if err != nil {
			continue
	}
	ctx, cancel := context.WithTimeout(context.Background(), rpcTimeout)
	client.Heartbeat(ctx, state)
	cancel()
	conn.Close()
	}
}

// this is the bully logic so when the leader is dead
// then the node with the lowest id will beocome a leader
// so it checks if there is a node wirh a lower id if they answer then that node will become the leader
//if the node does not answer then the node will become the leder
func (n *Node) startElection() {
	for id := 0; id < n.ID; id++ {
		if n.ping(id) != nil {
			fmt.Println("Node", n.ID, ": node", id, "is alive so it will become the leader")
			n.mu.Lock()
			n.lastHeartbeat = time.Now()
			n.mu.Unlock()
			return


		}
	
	}

	n.becomeLeader()
}

// this function is called when a node becomes the leader
// what it does is that it pings all the other nodes to get their state and then it applies the state to itself

func (n *Node) becomeLeader() {
	for id :=  range n.Peers {
		state := n.ping(id)


		n.mu.Lock()
		if state != nil && state.HighestBid > n.highestBid {
			n.applyState(state)
			
		}
		n.mu.Unlock()

	}

	n.mu.Lock()
	if n.endTime.IsZero() {
		n.endTime = time.Now().Add(auctionDuration)
	}
	n.isLeader = true
	n.leaderID = n.ID
	n.mu.Unlock()

	fmt.Println("Node", n.ID, "is now the LEADER")

}

// this function is used to ping a node to check if it is alive and get its state
// so basically it is used to get the state of the auction from the other nodes when a node becomes the leader
func (n *Node) ping(peerID int) *pb.State {
	client, conn, err := dial(n.Peers[peerID])
	if err != nil {
		return nil
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), rpcTimeout)
	defer cancel()

	state, err := client.Ping(ctx, &pb.Empty{})
	if err != nil {
		return nil
	}
	return state
}

func (s *ReplicationServer) Heartbeat(ctx context.Context, req *pb.State) (*pb.Empty, error) {
	n := s.node
	n.mu.Lock()
	defer n.mu.Unlock()

	n.lastHeartbeat = time.Now()
	n.leaderID = int(req.LeaderId)

	if n.endTime.IsZero() && req.EndTimeUnixMs != 0 {
		n.endTime = time.UnixMilli(req.EndTimeUnixMs)
	}
	// bids only go up, so a higher bid means newer state
	if req.HighestBid > n.highestBid {
		n.applyState(req)
	}
	return &pb.Empty{}, nil
}


func (s *ReplicationServer) Ping(ctx context.Context, req *pb.Empty) (*pb.State, error) {
	s.node.mu.Lock()
	defer s.node.mu.Unlock()
	return s.node.snapshot(), nil
}










