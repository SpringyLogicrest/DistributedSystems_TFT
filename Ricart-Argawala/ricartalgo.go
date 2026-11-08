package main

import (
	"fmt"
	"log"
	"sync"
)

// Requests access to the CS and returns an error if a request fails.
func RequestCriticalSectionAccess(node *Node) error {
	node.Mu.Lock()

	// Avoids starting another request while this node is already busy.
	if node.Requesting || node.InCS {
		node.Mu.Unlock()
		return fmt.Errorf("node %d is already requesting or inside the CS", node.ID)
	}

	node.clock++
	node.Requesting = true
	node.RequestTS = node.clock
	node.RepliesReceived = make(map[int64]bool)
	timestamp := node.RequestTS
	node.Mu.Unlock()

	fmt.Printf("Node %d: requesting access to critical section (timestamp=%d)\n", node.ID, timestamp)

	// here we send requests to all nodes in the network since the algo needs to send requests to all nodes and get a reply from all
	var wg sync.WaitGroup

	errors := make(chan error, len(node.Peers)) // Store request errors without stopping the program.

	for id, addr := range node.Peers {
		wg.Add(1)
		go func(id int64, addr string) {
			defer wg.Done()
			if err := SendRequest(node, id, addr, timestamp); err != nil {
				errors <- err
			}
		}(id, addr)
	}

	// so here we wait for all n-1 replies since we need approval from all to enter the cs
	wg.Wait()
	close(errors)

	// Cancel the request if we didn't receive all permissions.
	if err, failed := <-errors; failed {
		node.Mu.Lock()
		node.Requesting = false
		deferred := node.Deferred
		node.Deferred = nil
		node.Mu.Unlock()

		// Release any nodes that were waiting for our reply.
		for _, ch := range deferred {
			close(ch)
		}

		return err
	}

	node.Mu.Lock()
	node.Requesting = false
	node.InCS = true
	node.Mu.Unlock()

	log.Printf("Node %d: entering critical section", node.ID)

	return nil
}

func ExitCriticalSection(node *Node) {
	log.Printf("Node %d: is going to exit the critical section", node.ID)
	node.Mu.Lock()
	node.InCS = false
	deferred := node.Deferred
	node.Deferred = nil
	node.Mu.Unlock()

	for _, ch := range deferred {
		close(ch)
	}
}
