//this file makes it possible to send updates/information to backup nodes
package main

import (
	"context"
	"log"
	"sync"

	pb "SingleLeaderReplication/proto"
)

// this is the leader side 
func (n *Node) SendUpdate() {
	// Fix 1: send the whole state, not a string.
	
} 



	