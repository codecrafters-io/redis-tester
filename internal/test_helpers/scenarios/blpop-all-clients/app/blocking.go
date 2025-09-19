package main

import (
	"container/list"
	"net"
	"sync"
)

type BlockingManager struct {
	key     string
	waiters *list.List
	mu      sync.RWMutex
}

type Waiter struct {
	conn net.Conn
}

func NewBlockingManager() *BlockingManager {
	return &BlockingManager{
		waiters: list.New(),
	}
}

func (bm *BlockingManager) WaitForElement(key string, conn net.Conn) {
	bm.mu.Lock()
	bm.key = key
	bm.waiters.PushBack(conn)
	bm.mu.Unlock()
}

func (bm *BlockingManager) NotifyAllWaiters(key string, store *Store, resp *RESPCodec) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Only process if this key matches the one we're waiting for
	if bm.key != key {
		return
	}

	value := store.LPop(key)
	if value != nil && bm.waiters.Len() > 0 {
		// BUG: Send response to ALL clients instead of just the first one
		response := resp.EncodeArray([]string{key, *value})

		// Send to ALL waiting clients
		for element := bm.waiters.Front(); element != nil; element = element.Next() {
			conn := element.Value.(net.Conn)
			conn.Write(response)
		}

		// Clear all waiters since we've notified everyone
		bm.waiters = list.New()
		bm.key = ""
	}
}
