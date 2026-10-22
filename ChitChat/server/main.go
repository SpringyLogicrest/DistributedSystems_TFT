package main

//Join(),Leave(),Subscribe(),Publish()

import (
	pb "ChitChat/grpc"
	"context"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)

var clock Lamport

func main() {
	listener, err := net.Listen("tcp", ":50051")

	if err != nil {
		log.Fatal("Error eccured, ", err)
	}

	grpcServer := grpc.NewServer()

	chat := &chatServer{
		clients:      make(map[string]bool), //empty map for active clients
		logicalClock: 0,
	}

	pb.RegisterChatServiceServer(grpcServer, chat)

	log.Println("Chat server started on port 50051")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// Defines our server
type chatServer struct {
	pb.UnimplementedChatServiceServer

	mu           sync.Mutex      //ensures only one goroutine at a time can change our map/data
	clients      map[string]bool //used to see if clients are active. Active clients go here
	logicalClock int64
}

// Our Join() method that allows people to join the chatroom
func (s *chatServer) Join(
	ctx context.Context,
	req *pb.JoinRequest, //our JoinRequest from chat.proto
	//we return either an answer to our client or a potential error
) (*pb.JoinResponse, error) {

	clientID := req.GetClientId() //get the clients ID from the request and save it
	clock.Receive(req.GetLogicalTimestamp()) //update our logical clock with the clients timestamp

	s.mu.Lock() //protect data from concurrent changes

	s.clients[clientID] = true //register client as active
	s.mu.Unlock() 
	
	t := clock.Send()
	
	log.Println(clientID, " joined the chatroom!", "logical time ", t)

	return &pb.JoinResponse{ //Send a response back to client
		ClientId: clientID,
		LogicalTimestamp: t,
	}, nil
}
