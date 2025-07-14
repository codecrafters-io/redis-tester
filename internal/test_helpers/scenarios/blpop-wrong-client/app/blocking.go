package main

import (
	"container/list"
	"net"
	"sync"
)

type BlockingManager struct {
	waiters map[string]*list.List
	mu      sync.RWMutex
}

type Waiter struct {
	conn net.Conn
}

func NewBlockingManager() *BlockingManager {
	return &BlockingManager{
		waiters: make(map[string]*list.List),
	}
}

func (bm *BlockingManager) WaitForElement(key string, conn net.Conn) {
	waiter := &Waiter{
		conn: conn,
	}

	bm.mu.Lock()
	if bm.waiters[key] == nil {
		bm.waiters[key] = list.New()
	}
	bm.waiters[key].PushBack(waiter)
	bm.mu.Unlock()
}

func (bm *BlockingManager) NotifyAllWaiters(store *Store, resp *RESPCodec) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	for key, waiters := range bm.waiters {
		if waiters == nil || waiters.Len() == 0 {
			continue
		}
		value := store.LPop(key)
		if value != nil {
			element := waiters.Front()
			waiter := element.Value.(*Waiter)
			waiters.Remove(element)

			// if no more waiters, remove the key
			if waiters.Len() == 0 {
				delete(bm.waiters, key)
			}

			// send response directly to the waiter's connection
			response := resp.EncodeArray([]string{key, *value})
			waiter.conn.Write(response)
		}
	}
}
