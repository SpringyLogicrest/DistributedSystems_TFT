package main

//keeps track of Lamport clock, node ID and critical section

import "sync"

type Node struct {
	ID        int64
	clock     int64
	RequestTS int64

	Requesting bool
	InCS       bool

	Peers map[int64]string
	DeferredReplies map[int64]chan struct{}
	RepliesReceived map[int64]bool

	Mu sync.Mutex

	Deferred []chan struct{}
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

// Function to compare requests
func HasPriority(
	myTimestamp int64,
	myID int64,
	otherTimestamp int64,
	otherID int64,
) bool {
	if myTimestamp != otherTimestamp {
		return myTimestamp < otherTimestamp
	}
	return myID < otherID
}
