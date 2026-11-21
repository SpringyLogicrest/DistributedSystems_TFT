//this file makes it possible to send updates/information to backup nodes
package main

import (
	"context"
	"log"
	"time"

	pb "SingleLeaderReplication/proto"

	"google.golang.org/grpc"
)

// SendUpdate sends an update to all backup nodes.
func (n *Node) SendUpdate(update string) {
	func (n *Node) replicate(st *pb.State) {
	var wg sync.WaitGroup                       
	for _, pid := range n.alivePeers() {         
		wg.Add(1)
		go func(pid int) {                       
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), rpcTimeout)
			defer cancel()

			// reuse the client created once in connect() so we odnt need to create a new client for each replicaiton
			if _, err := n.replClients[pid].Replicate(ctx, st); err != nil {
				log.Printf("[node %d] replicate to %d failed: %v", n.ID, pid, err)
				n.markDead(pid)  
			}
		}(pid)
	}
	wg.Wait()
	}
}
	