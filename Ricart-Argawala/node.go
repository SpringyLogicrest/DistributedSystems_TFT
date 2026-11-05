package ricartargawala

//keeps track of Lamport clock, node ID and critical section

import "sync"

type Node struct {
	ID        int64
	clock     int64
	RequestTS int64

	Requesting bool
	InCS       bool

	Mu sync.Mutex
}
