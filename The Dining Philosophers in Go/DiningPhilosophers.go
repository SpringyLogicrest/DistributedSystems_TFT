package main

import "fmt"

// ForkRequest is the request that can be given to fork
type ForkRequest struct {
	philosopher string    //Name of philosopher that wishes to interact with fork.
	reply       chan bool //"Can I recieve a fork?" - will be answered by fork.
	finished    chan bool //"I am done with my fork" - will be answered by philosopher.
}

// Function for Fork that'll become a goroutine. Tells us what Fork does.
func Fork(requests chan ForkRequest) {
	for { //fork-goroutine can be used again and again
		req := <-requests //waiting for request
		fmt.Println("got request from " + req.philosopher)

		req.reply <- true //fork is now taken
		fmt.Println("fork taken by " + req.philosopher)

		<-req.finished //wait here until fork is given back

		fmt.Println("fork returned by " + req.philosopher)
	}
}

// Function for Philosopher that'll become a goroutine. Tells us what Philosopher does.
func Philosopher(name string, leftFork chan ForkRequest, rightFork chan ForkRequest) {
	fmt.Println(name + " is thinking...")

	//make a ForkRequest
	//make the channels from ForkRequest
	reply := make(chan bool)
	finished := make(chan bool)

	//the actual ForkRequest
	request := ForkRequest{
		philosopher: name, reply: reply, finished: finished,
	}

	//send request to left and right fork
	leftFork <- request
	rightFork <- request

	//wait for reply for both left and right fork
	<-reply
	<-reply

	fmt.Println(name + " is eating...")

	//philosopher is done eating
	finished <- true
	finished <- true

}

func main() {

	//fork1 is the channel that can communicate with the first fork (go Fork(fork1))
	fork1 := make(chan ForkRequest)
	fork2 := make(chan ForkRequest)
	fork3 := make(chan ForkRequest)
	fork4 := make(chan ForkRequest)
	fork5 := make(chan ForkRequest)

	//The forks as individual goroutines
	go Fork(fork1)
	go Fork(fork2)
	go Fork(fork3)
	go Fork(fork4)
	go Fork(fork5)

	//The philosophers as individual goroutines
	go Philosopher("Aedesia", fork1, fork2)
	go Philosopher("Hypatia", fork2, fork3)
	go Philosopher("Ghosha", fork3, fork4)
	go Philosopher("Arete", fork4, fork5)
	go Philosopher("Diotima", fork5, fork1)

}
