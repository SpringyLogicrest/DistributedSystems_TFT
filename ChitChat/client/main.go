package main

import (
	pb "ChitChat/grpc"
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var clock Lamport

func main() {
	// Get the client's ID from the terminal arguments.
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run ./client <ClientID>")
	}

	clientID := os.Args[1]

	// Create the logs directory if it doesn't already exist.
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatal("Could not create logs directory: ", err)
	}

	// Create a log file specifically for this client.
	logFile, err := os.OpenFile(
		filepath.Join("logs", "client_"+clientID+".log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)

	if err != nil {
		log.Fatal("Could not open client log file: ", err)
	}

	defer logFile.Close()

	// Create a logger that writes technical events to the log file.
	eventLog := log.New(logFile, "", log.LstdFlags)

	// Display a welcome message.
	fmt.Printf("Welcome to ChitChat, %s!\n", clientID)

	// Create a context that can be cancelled with Ctrl+C.
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
	)
	defer stop()

	conn, err := grpc.NewClient(
		"localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatal("Could not create connection: ", err) // Stops the program if creating the connection fails.
	}

	defer conn.Close() //connection closes when main() finishes

	client := pb.NewChatServiceClient(conn)

	// Subscribe to broadcasts from the server.
	stream, err := client.Subscribe(
		ctx,
		&pb.SubscribeRequest{
			ClientId:         clientID,
			LogicalTimestamp: clock.Send(),
		},
	)

	if err != nil {
		log.Fatal("Could not subscribe: ", err)
	}

	// Wait for the server to confirm that our subscription is registered.
	if _, err := stream.Header(); err != nil {
		log.Fatal("Subscription failed: ", err)
	}

	eventLog.Printf(
		"component=Client event=subscribed client_id=%s",
		clientID,
	)

	fmt.Println("Connected to the server!")

	// This goroutine continuously listens for broadcast messages from the server without blocking the rest of the client.
	go func() {
		for {
			// Wait for the next broadcast message.
			msg, err := stream.Recv()

			if err != nil {
				// If the stream closes unexpectedly, log the error and cancel the client's context.
				if ctx.Err() == nil {
					log.Printf(
						"component=Client event=stream_closed client_id=%s error=%v",
						clientID, err,
					)
					stop()
				}
				return
			}

			// Update the Lamport clock using the received timestamp.
			t := clock.Receive(msg.GetLogicalTimestamp())

			// Log the full event for documentation and debugging.
			eventLog.Printf(
				"component=Client event=message_received client_id=%s message_time=%d local_time=%d content=%q",
				clientID,
				msg.GetLogicalTimestamp(),
				t,
				msg.GetContent(),
			)

			// Display the broadcast in a user-friendly format.
			fmt.Printf(
				"[L:%d] %s\n",
				msg.GetLogicalTimestamp(),
				msg.GetContent(),
			)
		}
	}()

	//Client joins chatroom
	response, err := client.Join(
		ctx,
		&pb.JoinRequest{
			ClientId:         clientID,
			LogicalTimestamp: clock.Send(),
		},
	)

	if err != nil {
		log.Fatal("Could not join ChitChat: ", err)
	}
	t := clock.Receive(response.GetLogicalTimestamp())
	eventLog.Printf(
		"component=Client event=joined client_id=%s logical_time=%d",
		response.GetClientId(),
		t,
	)

	fmt.Printf("You have joined ChitChat as %s!\n", response.GetClientId())

	// Keep the client connected until Ctrl+C is pressed or the subscription stream closes.
	log.Println("Listening for messages. Press Ctrl+C to exit.")

	//Client will be able to send messages through the terminal
	go func() {
		scanner := bufio.NewScanner(os.Stdin) //allows for clients to write chat messages

		for scanner.Scan() {
			message := scanner.Text()

			publishResponse, err := client.Publish(
				ctx,
				&pb.PublishRequest{
					ClientId:         clientID,
					Content:          message,
					LogicalTimestamp: clock.Send(),
				},
			)

			if err != nil {
				log.Println("Coould not publish message:", err)
			} else {
				t := clock.Receive(publishResponse.GetLogicalTimestamp())
				log.Printf("message sent at logical time: %d", t)

				// Log the event
				eventLog.Printf(
					"component=Client event=publishing client_id=%s logical_time=%d",
					clientID,
					t,
				)
			}
		}
	}()

	<-ctx.Done()

	//Tell server client is leaving
	leaveResponse, err := client.Leave(
		context.Background(), &pb.LeaveRequest{
			ClientId:         clientID,
			LogicalTimestamp: clock.Send(),
		},
	)
	if err != nil {
		log.Fatal("Could not leave ChitChat: ", err)
	} else {
		fmt.Println("You have left ChitChat. See you another time!")
	}

	//Update Lamport clock
	t = clock.Receive(leaveResponse.GetLogicalTimestamp())

	// Log the leave event
	eventLog.Printf(
		"component=Client event=left client_id=%s logical_time=%d",
		clientID,
		t,
	)

}
