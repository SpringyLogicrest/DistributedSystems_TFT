package main

import ("fmt"
		"sync"
		"log")


func RequestCriticalSectionAccess(node *Node) {
	node.Mu.Lock()
	node.clock++
	node.Requesting = true
	node.RequestTS = node.clock
	node.RepliesReceived = make(map[int64]bool)
	node.Mu.Unlock()
	
	fmt.Printf("Node %d: requesting access to critical section (timestamp=%d)\n", node.ID, node.RequestTS)


	// here we send requests to all nodes in the network since the algo needs to send requests to all nodes and get a reply from all
	var wg sync.WaitGroup
	for id, addr := range node.Peers {
		wg.Add(1)
		go func(id int64, addr string) {
			defer wg.Done()
			if err := SendRequest(node, id, addr, node.RequestTS); err != nil {
				log.Fatalf("Node %d: %v", node.ID, err)
			}
		}(id, addr)
	}

	// so here we wait for all n-1 replies since we need approval from all to enter the cs 
	wg.Wait()

	node.Mu.Lock()
	node.Requesting = false
	node.InCS = true
	node.Mu.Unlock()
	
	fmt.Printf("Node %d: entering critical section\n", node.ID)

}

func ExitCriticalSection(node *Node) {
	fmt.Printf("Node %d: is going to exit the critical section\n", node.ID)
	node.Mu.Lock()
	node.InCS = false
	deferred := node.Deferred
	node.Deferred = nil
	node.Mu.Unlock()
		
	for _, ch := range deferred {
		close(ch)
	}
}


