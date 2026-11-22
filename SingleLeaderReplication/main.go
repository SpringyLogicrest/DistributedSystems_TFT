package main

import (
	"flag"
	"fmt"
	"net"

	"time"

	pb "SingleLeaderReplication/proto"

	"google.golang.org/grpc"
)

var addrs = []string{"localhost:5000", "localhost:5001", "localhost:5002"}

func main() {
	id := flag.Int("id", 0, "ID of this node")
	flag.Parse()

	if *id < 0 || *id >= len(addrs) {
		fmt.Println("Invalid node ID")
		return
	}

	node := NewNode(*id, 0)

	for peerID, address := range addrs {
		if peerID != node.ID {
			node.Peers[peerID] = address
		}
	}

	fmt.Println("Starting node:", node.ID)
	fmt.Println("Is leader:", node.isLeader)

	address := addrs[*id]

	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Failed to listen:", err)
		return
	}

	grpcServer := grpc.NewServer()

	auctionServer := &AuctionServer{
		node: node,
	}

	replicationServer := &ReplicationServer{
		node: node,
	}

	pb.RegisterReplicationServiceServer(grpcServer, replicationServer)

	pb.RegisterAuctionServiceServer(grpcServer, auctionServer)

	node.lastHeartbeat = time.Now().Add(startupDelay)

	
	if node.isLeader {
		node.endTime = time.Now().Add(auctionDuration)
	}

	
	go node.RunFailureDetection()

	fmt.Println("Node", node.ID, "listening on", address)

	if err := grpcServer.Serve(listener); err != nil {
		fmt.Println("Server failed:", err)
	}

	

}
