package main

import (
	pb "ChitChat/grpc"
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var clock Lamport

func main() {
	conn, err := grpc.NewClient(
		"localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Println("Could not create connection: ", err)
	}

	defer conn.Close() //connection closes when main() finishes

	client := pb.NewChatServiceClient(conn)

	response, err := client.Join(
		context.Background(),
		&pb.JoinRequest{ClientId: "ClientA", LogicalTimestamp: clock.Send()},
	)

	if err != nil {
		log.Fatal("Could not join chat: ", err)
	}
	t := clock.Receive(response.GetLogicalTimestamp())
	log.Println("Joined chatroom as", response.GetClientId(), "with logical timestamp", t)
}
