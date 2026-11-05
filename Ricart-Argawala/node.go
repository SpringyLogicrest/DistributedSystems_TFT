package main

//keeps track of Lamport clock, node ID and critical section

import "sync"

type Node struct {
	ID        int64
	clock     int64
	RequestTS int64

	Requesting bool
	InCS       bool

	DeferredReplies []int64
	RepliesReceived map[int64]bool

	Mu sync.Mutex
}

// Increases the Lamport clock before sending a message.
func (n *Node) NextTimestamp() int64 {
	n.Mu.Lock()
	defer n.Mu.Unlock()

	n.clock++
	return n.clock
}

// Updates the Lamport clock when receiving a message.
func (n *Node) ReceiveTimestamp(timestamp int64) int64 {
	n.Mu.Lock()
	defer n.Mu.Unlock()

	if timestamp > n.clock {
		n.clock = timestamp
	}

	n.clock++
	return n.clock
}
