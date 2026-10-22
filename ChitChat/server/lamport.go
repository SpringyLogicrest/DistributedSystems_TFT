// here i just a fyi, in lamport clock whenever a event occur we increment the time by 1

package main

import (
	"sync"
)

type Lamport struct {
	mu sync.Mutex
	time int64
}

// here we just send the time and increment it by 1, just like the lamport clock does
// also another fyi we use the mutex lock to make sure only one goroutine can get the time.
// this prevents race conditions and makes sure the time is always correct.
func (c *Lamport) Send() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.time++
	return c.time
}

// here we receive the time and increment it by 1, just like the lamport clock does

func (c *Lamport) Receive(receivedTime int64) int64{
	c.mu.Lock()
	defer c.mu.Unlock()
	if receivedTime > c.time {
		c.time = receivedTime
	}
	return c.time
}
