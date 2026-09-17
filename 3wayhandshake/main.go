package main

//3 way tcp handshake
import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var wg sync.WaitGroup

func server(syn <-chan string, ack <-chan string, synack chan<- string) {
	// this tells the waitgroup that the server goroutine is done when it returns
	// the defer makes it like always run even if there is a error.
	defer wg.Done()

	msg := <-syn // Wait for initial SYN
	fmt.Println("Server has got the message:", msg)

	synack <- "SYN-ACK" // Send initial SYN-ACK
	fmt.Println("Server has sent the message: SYN-ACK")

	for {
		select {
		case msg := <-ack:
			// Server has recieved the ACK, and can consider the handshake complete
			fmt.Println("Server has got the message:", msg)
			fmt.Println("Server: Handshake completed")
			return

		case msg := <-syn:
			// Client sent SYN again, probably because SYN-ACK was lost
			fmt.Println("Server has got another:", msg)
			fmt.Println("Server assumes previous SYN-ACK was lost")

			synack <- "SYN-ACK"
			fmt.Println("Server has resent the message: SYN-ACK")

		case <-time.After(2 * time.Second):
			// ACK may have been lost
			fmt.Println("Server timed out waiting for ACK. Resending SYN-ACK...")

			synack <- "SYN-ACK"
			fmt.Println("Server has resent the message: SYN-ACK")
		}
	}
}

func client(syn chan<- string, ack chan<- string, synack <-chan string) {
	defer wg.Done()

	for {
		syn <- "SYN" //Client -> server
		fmt.Println("Client has sent the message: SYN")

		select {
		case <-synack: //Client waits for SYN-ACK
			fmt.Println("Client has got the message: SYN-ACK")

			ack <- "ACK" //Client -> server
			fmt.Println("Client has sent the message: ACK")

			// Stay alive briefly in case the ACK gets lost
			for {
				select {
				case <-synack:
					fmt.Println("Client received SYN-ACK again")
					fmt.Println("Client assumes previous ACK was lost")

					ack <- "ACK"
					fmt.Println("Client has resent the message: ACK")

				case <-time.After(3 * time.Second):
					fmt.Println("Client assumes handshake is complete")
					return
				}
			}

		case <-time.After(2 * time.Second):
			fmt.Println("Client timed out waiting for SYN-ACK. Sending SYN again...")
		}
	}
}

// Forwarder is meant to simulate an unreliable network by delaying or dropping messages
// All messages between the client and server must pass through the forwarder (done in main)
// Also, as a reminder (more to myself, aka Freddy) forwarder can only recieve from input, and can only send to output
func forwarder(input <-chan string, output chan<- string, lossRate float64) {
	for msg := range input {
		fmt.Println("Forwarder received:", msg)

		randomDelay := time.Duration(rand.Intn(1000)) * time.Millisecond // Randomized delay
		fmt.Println("Forwarder delaying for: ", randomDelay)

		time.Sleep(randomDelay)

		if rand.Float64() < lossRate {
			fmt.Println("Forwarder dropped: ", msg)
			continue
		}

		fmt.Println("Forwarder forwarded: ", msg)
		output <- msg
	}
}

func main() {
	// Client -> Forwarder -> Server
	clientSyn := make(chan string)
	serverSyn := make(chan string)

	clientAck := make(chan string)
	serverAck := make(chan string)

	// Server -> Forwarder -> Client
	serverSynack := make(chan string)
	clientSynack := make(chan string)

	// 20% chance of losing a message
	lossRate := 0.2

	// SYN goes Client -> Forwarder -> Server
	go forwarder(clientSyn, serverSyn, lossRate)

	// ACK goes Client -> Forwarder -> Server
	go forwarder(clientAck, serverAck, lossRate)

	// SYN-ACK goes Server -> Forwarder -> Client
	go forwarder(serverSynack, clientSynack, lossRate)

	wg.Add(2)

	go client(clientSyn, clientAck, clientSynack)
	go server(serverSyn, serverAck, serverSynack)

	wg.Wait()
}
