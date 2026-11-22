// client can send bids and result
package main

import (
	pb "SingleLeaderReplication/proto"
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var addrs = []string{"localhost:5000", "localhost:5001", "localhost:5002"}

// connect creates a client for one node
func connect(address string) (pb.AuctionServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return pb.NewAuctionServiceClient(conn), conn, nil
}

// here we have to sendbid to the leader node and if the leader is dead we send it to the new leader node (after the election is done)
func sendBid(bidderID int64, amount int64) string {
	request := &pb.BidRequest{Amount: amount, BidderId: bidderID}

	for attempt := 1; attempt <= 5; attempt++ {
		for _, address := range addrs {
			client, conn, err := connect(address)
			if err != nil {
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			response, err := client.Bid(ctx, request)
			cancel()
			conn.Close()

			if err != nil {
				// nthis checks if a node is deaaaaaaad
				continue
			}
			if response.Ack == "exception" {
				// if a node is not leader go next 
				continue
			}

			fmt.Println("(answered by", address+")")
			return response.Ack 
		}

		fmt.Println("No leader found, waiting for election... (attempt", attempt, "of 5)")
		time.Sleep(1 * time.Second)
	}
	return "exception"
}

// this is the function that gets the result of the auction from any node (leader or backup)
func getResult() (*pb.ResultResponse, error) {
	for _, address := range addrs {
		client, conn, err := connect(address)
		if err != nil {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		response, err := client.Result(ctx, &pb.ResultRequest{})
		cancel()
		conn.Close()

		if err == nil {
			return response, nil
		}
	}
	return nil, fmt.Errorf("no node is reachable")
}

func main() {
	var bidderID int64

	fmt.Print("Choose your bidder ID: ")
	if _, err := fmt.Scan(&bidderID); err != nil {
		fmt.Println("Invalid bidder ID:", err)
		return
	}

	for {
		fmt.Print("\n1) Bid   2) Result   3) Quit\nChoose: ")
		var choice int
		if _, err := fmt.Scan(&choice); err != nil {
			fmt.Println("Invalid choice")
			return
		}

		switch choice {
		case 1:
			var amount int64
			fmt.Print("How much do you want to bid?: ")
			if _, err := fmt.Scan(&amount); err != nil {
				fmt.Println("Invalid bid:", err)
				continue
			}
			ack := sendBid(bidderID, amount)
			fmt.Println("Server response:", ack)

		case 2:
			result, err := getResult()
			if err != nil {
				fmt.Println("Failed to get result:", err)
				continue
			}
			if result.AuctionEnded {
				fmt.Println("Auction is OVER. Winner: bidder", result.HighestBidder, "with", result.HighestBid)
			} else {
				fmt.Println("Auction is running. Highest bid:", result.HighestBid, "by bidder", result.HighestBidder)
			}


			// this is just a quitter for the client
		case 3: 
			return
		}
	}
}