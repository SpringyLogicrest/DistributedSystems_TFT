package main

//3 way tcp handshake
import (
	"fmt"
	"sync"
)

var wg sync.WaitGroup

func server(syn <-chan string, ack <-chan string, synack chan<- string) {
	// this tells the waitgroup that the server goroutine is done when it returns
	// the defer makes it like always run even if there is a error.
	defer wg.Done()

	msg := <-syn
	fmt.Println("Server has got the message: ", msg)

	synack <- "SYN-ACK"
	fmt.Println("Server has sent the message: SYN-ACK")

	msg2 := <-ack
	fmt.Println("Server has got the message: ", msg2)
}

func client(syn chan<- string, ack chan<- string, synack <-chan string) {
	defer wg.Done()

	syn <- "SYN" //Client -> server
	fmt.Println("Client has sent the message: SYN")

	<-synack //Client waits for SYN-ACK
	fmt.Println("Client has got the message: SYN-ACK")

	ack <- "ACK" //Client -> server
	fmt.Println("Client has sent the message: ACK")
}

func main() {
	syn := make(chan string)
	ack := make(chan string)
	synack := make(chan string)

	wg.Add(2)

	go client(syn, ack, synack)
	go server(syn, ack, synack)

	wg.Wait()
}
