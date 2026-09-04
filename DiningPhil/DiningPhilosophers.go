package main

import (
	"fmt"
	"math/rand"
	"time"
)

// ForkRequest is the request that can be given to fork
type ForkRequest struct {
	philosopher string    //Name of philosopher that wishes to interact with fork.
	reply       chan bool //"Can I recieve a fork?" - will be answered by fork.
	finished    chan bool //"I am done with my fork" - will be answered by philosopher.
}

// MealUpdate is sent whenever a philosopher finishes a meal.
type MealUpdate struct {
	philosopher string
	count       int // How many meals has a specific philosopher eaten?
}

// Logger prints messages one at a time.
// This is done in order to have a less messy output (not a requirement, but it just looks nice in my opinion)
func Logger(logs chan string) {
	for {
		message := <-logs
		fmt.Println(message)
	}
}

// Monitor keeps track of how many times each philosopher has eaten.
func Monitor(mealUpdates chan MealUpdate, logs chan string) {
	mealCounts := make(map[string]int)

	allReachedThree := false

	for {
		update := <-mealUpdates

		mealCounts[update.philosopher] = update.count

		// Print a message once everyone has eaten at least 3 times
		if !allReachedThree {
			if mealCounts["Aedesia"] >= 3 && mealCounts["Hypatia"] >= 3 && mealCounts["Ghosha"] >= 3 && mealCounts["Arete"] >= 3 && mealCounts["Diotima"] >= 3 {

				allReachedThree = true

				logs <- `
				======================================================
				     EVERY PHILOSOPHER HAS EATEN AT LEAST 3 TIMES!
				======================================================`
			}
		}
	}
}

// Function for Fork that'll become a goroutine. Tells us what Fork does.
func Fork(name string, requests chan ForkRequest, logs chan string) {
	for { //fork-goroutine can be used again and again
		req := <-requests //waiting for request
		logs <- fmt.Sprintf("[FORK %s] Request received from %s", name, req.philosopher)

		req.reply <- true //fork is now taken by the philosopher
		logs <- fmt.Sprintf("[FORK %s] Taken by %s", name, req.philosopher)

		<-req.finished // Wait until the philosopher returns the fork.

		logs <- fmt.Sprintf("[FORK %s] Returned by %s", name, req.philosopher)
	}
}

// Function for Philosopher that'll become a goroutine. Tells us what Philosopher does.
func Philosopher(name string, leftFork chan ForkRequest, rightFork chan ForkRequest, reverseOrder bool, mealUpdates chan MealUpdate, logs chan string) {
	fmt.Println(name + " is thinking...")

	// Make the channels used in ForkRequest.
	reply := make(chan bool)
	finished := make(chan bool)

	// Keep track of how many times this philosopher has eaten.
	eaten := 0

	for {
		// ===== THINKING LOGIC ===== //
		// here is the thinking time (inspired by sharingtoolbox)
		//basically we take a random number between 1 and 5 (or you can use any time) and then sleep for that amount of time
		think_time := rand.Intn(5) + 1 //random time to think
		logs <- fmt.Sprintf("[%-7s] THINKING | %d seconds", name, think_time)

		time.Sleep(time.Duration(think_time) * time.Second)

		//the actual ForkRequest
		request := ForkRequest{
			philosopher: name,
			reply:       reply,
			finished:    finished,
		}

		// ===== GET THE FORKS ORDER ===== //
		if reverseOrder {
			// Requests the right fork first.
			rightFork <- request
			<-reply // Wait until the right fork has actually been acquired.

			leftFork <- request
			<-reply // Wait until the left fork has actually been acquired.
		} else {
			// Requests the left fork first.
			leftFork <- request
			<-reply // Wait until the left fork has actually been acquired.

			rightFork <- request
			<-reply // Wait until the right fork has actually been acquired.
		}

		// ===== EATING LOGIC ===== //
		// the exact same from previous sleeping time just with a different variable name for eating,
		eat_time := rand.Intn(5) + 1 //random time to eat
		logs <- fmt.Sprintf("[%-7s] EATING   | %d seconds", name, eat_time)
		time.Sleep(time.Duration(eat_time) * time.Second)

		// Increase and display the number of times this philosopher has eaten (since all philosophers have to eat at least thrice).
		eaten++

		logs <- fmt.Sprintf("[%-7s] FINISHED | Meal #%d", name, eaten)

		// Tell the monitor that this philosopher has eaten.
		mealUpdates <- MealUpdate{
			philosopher: name,
			count:       eaten,
		}

		// ===== RETURN FORKS ===== //
		finished <- true
		finished <- true

		// The loop starts again and the philosopher goes back to thinking.
	}

}

func main() {
	// Each channel communicates with one individual fork (so fork1 is the first fork).
	fork1 := make(chan ForkRequest)
	fork2 := make(chan ForkRequest)
	fork3 := make(chan ForkRequest)
	fork4 := make(chan ForkRequest)
	fork5 := make(chan ForkRequest)

	// Philosophers report completed meals through this channel.
	mealUpdates := make(chan MealUpdate)

	// All goroutines send their output through this channel.
	logs := make(chan string)

	go Logger(logs)
	go Monitor(mealUpdates, logs)

	//The forks as individual goroutines
	go Fork("1", fork1, logs)
	go Fork("2", fork2, logs)
	go Fork("3", fork3, logs)
	go Fork("4", fork4, logs)
	go Fork("5", fork5, logs)

	//The philosophers as individual goroutines
	// false = left fork first
	// true  = right fork first
	go Philosopher("Aedesia", fork1, fork2, false, mealUpdates, logs)
	go Philosopher("Hypatia", fork2, fork3, false, mealUpdates, logs)
	go Philosopher("Ghosha", fork3, fork4, false, mealUpdates, logs)
	go Philosopher("Arete", fork4, fork5, false, mealUpdates, logs)
	go Philosopher("Diotima", fork5, fork1, true, mealUpdates, logs) // Reverse order to prevent a potential deadlock

	// Keeps the main alive
	select {}
}

/*
	DEADLOCK PREVENTION:
		If every philosopher took their left fork first, it would be possible for all five philosophers to hold one fork while waiting forever for the next fork.

		So we decided to make it so an n number of philosophers takes their forks in the opposite order.
		This results in the circular waiting condition needed for a deadlock cannot be fulfilled, and therefore preventing a deadlock!
*/
