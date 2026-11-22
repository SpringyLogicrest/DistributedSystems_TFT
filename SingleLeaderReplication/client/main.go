// client can send bids!
package main

import (
	pb "SingleLeaderReplication/proto"
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient(
		"localhost:5000",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		fmt.Println("Failed to create connection:", err)
		return
	}

	defer conn.Close()

	client := pb.NewAuctionServiceClient(conn)

	var bidderID int64
	var amount int64

	fmt.Print("Choose your bidder ID: ")
	if _, err := fmt.Scan(&bidderID); err != nil {
		fmt.Println("Invalid bidder ID:", err)
		return
	}

	fmt.Print("How much do you want to bid?: ")
	if _, err := fmt.Scan(&amount); err != nil {
		fmt.Println("Invalid bid:", err)
		return
	}

	request := &pb.BidRequest{
		Amount:   amount,
		BidderId: bidderID,
	}

	//Send bid to leader/Node 0
	ctx, cancel := context.WithTimeout( //5 sec to answer
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	response, err := client.Bid(ctx, request)

	if err != nil {
		fmt.Println("Failed to send bid:", err)
		return
	}

	fmt.Println("Server response:", response.Ack)

	fmt.Println("Bidder Id: ", bidderID)
	fmt.Println("You bidded: ", amount)

}
