package main

//Join(),Leave(),Subscribe(),Publish()

import (
	pb "ChitChat/grpc"
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var clock Lamport

func main() {
	listener, err := net.Listen("tcp", ":50051")

	if err != nil {
		log.Fatal("Error eccured, ", err)
	}

	grpcServer := grpc.NewServer()

	chat := &chatServer{
		clients:     make(map[string]bool), //empty map for active clients
		subscribers: make(map[string]chan *pb.BroadcastMessage),
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

	mu sync.Mutex //ensures only one goroutine at a time can change our map/data

	clients     map[string]bool                      //used to see if clients are active. Active clients go here
	subscribers map[string]chan *pb.BroadcastMessage // Maps each subscribed client's ID to a channel used to deliver broadcast messages to that client.
}

// Our Join() method that allows people to join the chatroom
func (s *chatServer) Join(
	ctx context.Context,
	req *pb.JoinRequest, //our JoinRequest from chat.proto
	//we return either an answer to our client or a potential error
) (*pb.JoinResponse, error) {

	clientID := req.GetClientId()            //get the clients ID from the request and save it
	clock.Receive(req.GetLogicalTimestamp()) //update our logical clock with the clients timestamp

	s.mu.Lock() //protect data from concurrent changes

	s.clients[clientID] = true //register client as active

	t := clock.Send()

	// Create a broadcast message to notify all subscribed clients that a new participant has joined the chatroom.
	message := &pb.BroadcastMessage{
		Content: fmt.Sprintf(
			"Participant %s joined Chit Chat at logical time %d",
			clientID, t,
		),
		LogicalTimestamp: t,
	}

	// Unlock the mutex before broadcasting to avoid locking the server twice when calling broadcast().
	s.mu.Unlock()

	s.broadcast(message)

	// Log the join event with the client ID and logical timestamp.
	log.Printf(
		"component=Server event=join client_id=%s logical_time=%d",
		clientID, t,
	)

	return &pb.JoinResponse{ // Send a response back to client
		ClientId:         clientID,
		LogicalTimestamp: t,
	}, nil
}

// Our Subscribe() method, which'll allow clients to continuously receive broadcast messages from the server using gRPC server-side streaming.
// Each client gets its own buffered channel, and the method keeps the stream open until the client disconnects.
func (s *chatServer) Subscribe(
	req *pb.SubscribeRequest,
	stream grpc.ServerStreamingServer[pb.BroadcastMessage],
) error {

	clientID := req.GetClientId()

	// Update Lamport clock when receiving the request.
	clock.Receive(req.GetLogicalTimestamp())

	// Every subscriber gets its own buffered channel.
	messageChannel := make(chan *pb.BroadcastMessage, 32)

	s.mu.Lock()

	// Prevent multiple subscriptions using the same ID.
	if _, exists := s.subscribers[clientID]; exists {
		s.mu.Unlock()
		return status.Error(
			codes.AlreadyExists,
			"client is already subscribed",
		)
	}

	s.subscribers[clientID] = messageChannel
	s.mu.Unlock()

	log.Printf(
		"component=Server event=subscribe client_id=%s",
		clientID,
	)

	// Signal that the subscription is registered.
	if err := stream.SendHeader(metadata.MD{}); err != nil {
		s.removeSubscriber(clientID)
		return err
	}

	// Clean up when this stream ends.
	defer s.removeSubscriber(clientID)

	for {
		select {

		case <-stream.Context().Done():
			// Client disconnected or cancelled subscription.
			return nil

		case msg := <-messageChannel:
			// Send the message to the client over gRPC.
			if err := stream.Send(msg); err != nil {
				return err
			}

			log.Printf(
				"component=Server event=message_delivery client_id=%s logical_time=%d",
				clientID,
				msg.GetLogicalTimestamp(),
			)
		}
	}
}

// Helper method to prevent the server from continuing to treat a disconnected client as active.
func (s *chatServer) removeSubscriber(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.subscribers, clientID)
	delete(s.clients, clientID)

	log.Printf(
		"component=Server event=unsubscribe client_id=%s",
		clientID,
	)
}

// Our broadcast() method, which'll go through all the subscribed clients and queues the message for delivery.
func (s *chatServer) broadcast(
	msg *pb.BroadcastMessage,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for clientID, messageChannel := range s.subscribers {

		// Don't allow one full channel to block everyone.
		select {
		case messageChannel <- msg:

		default:
			log.Printf(
				"component=Server event=queue_full client_id=%s",
				clientID,
			)
		}
	}

	log.Printf(
		"component=Server event=broadcast logical_time=%d content=%q",
		msg.GetLogicalTimestamp(),
		msg.GetContent(),
	)
}
