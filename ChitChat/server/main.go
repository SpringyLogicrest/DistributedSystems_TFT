package main

//Join(),Leave(),Subscribe(),Publish()

import (
	pb "ChitChat/grpc"
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"unicode/utf8"

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
			"Participant %s joined ChitChat at logical time %d",
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

// Our Leave() method, which'll allow clients to leave the chatroom.
func (s *chatServer) Leave(
	ctx context.Context, //information about GRPC call
	req *pb.LeaveRequest, //our LeaveRequest from chat.proto
) (*pb.LeaveResponse, error) {

	clientID := req.GetClientId()            //get the clients ID from the request and save it
	clock.Receive(req.GetLogicalTimestamp()) //update our logical clock with the clients timestamp

	s.mu.Lock() //protect data from concurrent changes

	t := clock.Send()

	//create broadcast message and let all active client know that the user has left the chatroom
	message := &pb.BroadcastMessage{
		Content: fmt.Sprintf(
			"Participant %s left ChitChat at logical time %d", clientID, t,
		), LogicalTimestamp: t,
	}

	//unlock mutex before broadcasting
	s.mu.Unlock()

	//Client no longer subscribes to recieve messages
	s.removeSubscriber(clientID)

	//broadcast message to everyone
	s.broadcast(message)

	// Log the join event with the client ID and logical timestamp.
	log.Printf(
		"component=Server event=leave client_id=%s logical_time=%d",
		clientID, t,
	)

	return &pb.LeaveResponse{LogicalTimestamp: t}, nil //return a Leave response
}

// publish method allows clients to send chatmessages
func (s *chatServer) Publish(
	ctx context.Context,
	req *pb.PublishRequest,
) (*pb.PublishResponse, error) {

	clientID := req.GetClientId() //get the clients ID from the request and save it
	content := req.GetContent()   //get message from client

	// Check that the message is valid UTF-8
	if !utf8.ValidString(content) {
		return nil, status.Error(
			codes.InvalidArgument,
			"message must be valid UTF-8",
		)
	}

	// Check message length doesnt exceed 128 characters
	if utf8.RuneCountInString(content) > 128 {
		return nil, status.Error(
			codes.InvalidArgument,
			"message exceeds 128 characters",
		)
	}

	clock.Receive(req.GetLogicalTimestamp()) //update our logical clock with the clients timestamp
	s.mu.Lock()                              //protect data from concurrent changes

	if !s.clients[clientID] {
		s.mu.Unlock()
		return nil, status.Error(
			codes.PermissionDenied,
			"client is not active",
		)
	}

	t := clock.Send()

	//create broadcast message
	message := &pb.BroadcastMessage{
		Content:          fmt.Sprintf("%s: %s", clientID, content),
		LogicalTimestamp: t,
	}

	s.mu.Unlock()

	//broadcast message to everyone
	s.broadcast(message)

	// Log the publish event with the client ID and logical timestamp.
	log.Printf(
		"component=Server event=publish client_id=%s logical_time=%d",
		clientID, t,
	)

	return &pb.PublishResponse{LogicalTimestamp: t}, nil
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
