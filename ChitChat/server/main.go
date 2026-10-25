package main

//Join(),Leave(),Subscribe(),Publish()

import (
	pb "ChitChat/grpc"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"unicode/utf8"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var clock Lamport

func main() {
	// Create a folder for our logs if it doesn't exists.
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatal(err)
	}

	// Create the server log file.
	logFile, err := os.OpenFile(
		"logs/server.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	// Log to both the terminal and the file.
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

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

	log.Println("component=Server event=startup port=50051")

	// Run the server in a goroutine.
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Printf("component=Server event=serve_error error=%v", err)
		}
	}()

	// Listen for Ctrl+C or a termination signal.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	<-signals

	log.Println("component=Server event=shutdown_requested")

	// Stop gRPC and close active connections.
	grpcServer.Stop()

	log.Println("component=Server event=shutdown")
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

	clientID := req.GetClientId() //get the clients ID from the request and save it
	s.mu.Lock()                   //protect data from concurrent changes

	// Check if the client has already joined the chatroom.
	if s.clients[clientID] {
		s.mu.Unlock()
		return nil, status.Error(
			codes.AlreadyExists,
			"client has already joined",
		)
	}

	clock.Receive(req.GetLogicalTimestamp()) //update our logical clock with the clients timestamp

	if !s.canBroadcastLocked() {
		s.mu.Unlock()
		return nil, status.Error(
			codes.ResourceExhausted,
			"a client's message queue is full",
		)
	}

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
	s.broadcastLocked(message)
	s.mu.Unlock()

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

	clientID := req.GetClientId() //get the clients ID from the request and save it
	s.mu.Lock()                   //protect data from concurrent changes

	// Check whether the client is actually in the chatroom.
	if !s.clients[clientID] {
		s.mu.Unlock()
		return nil, status.Error(
			codes.NotFound,
			"client is not active",
		)
	}

	clock.Receive(req.GetLogicalTimestamp()) //update our logical clock with the clients timestamp

	t := clock.Send()

	//create broadcast message and let all active client know that the user has left the chatroom
	message := &pb.BroadcastMessage{
		Content: fmt.Sprintf(
			"Participant %s left Chit Chat at logical time %d", clientID, t,
		), LogicalTimestamp: t,
	}

	// Remove the leaving client before notifying everyone else.
	delete(s.clients, clientID)
	delete(s.subscribers, clientID)

	// Broadcast while still holding the lock.
	s.broadcastLocked(message)
	s.mu.Unlock()

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

	s.mu.Lock()                              //protect data from concurrent changes
	clock.Receive(req.GetLogicalTimestamp()) //update our logical clock with the clients timestamp

	if !s.canBroadcastLocked() {
		s.mu.Unlock()
		return nil, status.Error(
			codes.ResourceExhausted,
			"a client's message queue is full",
		)
	}

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

	//broadcast message to everyone
	s.broadcastLocked(message)
	s.mu.Unlock()

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

	// Every subscriber gets its own buffered channel.
	messageChannel := make(chan *pb.BroadcastMessage, 32)

	s.mu.Lock()
	clock.Receive(req.GetLogicalTimestamp()) // Update Lamport clock when receiving the request.

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

	// The client may already have left through Leave().
	if _, exists := s.subscribers[clientID]; !exists {
		return
	}

	wasActive := s.clients[clientID]

	delete(s.subscribers, clientID)
	delete(s.clients, clientID)

	// Notify other participants if an active client disconnects.
	if wasActive {
		t := clock.Send()

		message := &pb.BroadcastMessage{
			Content: fmt.Sprintf(
				"Participant %s left Chit Chat at logical time %d",
				clientID, t,
			),
			LogicalTimestamp: t,
		}

		s.broadcastLocked(message)
	}

	log.Printf(
		"component=Server event=unsubscribe client_id=%s",
		clientID,
	)
}

// Our broadcast() method, which'll go through all the subscribed clients and queues the message for delivery.
func (s *chatServer) broadcastLocked( //
	msg *pb.BroadcastMessage,
) {
	for clientID, messageChannel := range s.subscribers {
		if !s.clients[clientID] {
			continue
		}

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

// Check whether every active subscriber has queue space.
// The caller must hold s.mu.
func (s *chatServer) canBroadcastLocked() bool {
	for id, ch := range s.subscribers {
		if s.clients[id] && len(ch) == cap(ch) {
			return false
		}
	}
	return true
}
