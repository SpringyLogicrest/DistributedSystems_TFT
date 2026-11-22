package main

import (
	"flag"
	"fmt"
	"net"

	//"time"

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

	pb.RegisterAuctionServiceServer(grpcServer, auctionServer)

	fmt.Println("Node", node.ID, "listening on", address)

	if err := grpcServer.Serve(listener); err != nil {
		fmt.Println("Server failed:", err)
	}

}
