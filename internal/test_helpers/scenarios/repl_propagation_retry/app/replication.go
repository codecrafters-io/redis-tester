package main

import (
	"net"
	"sync"
)

// ReplicationManager handles replication state and operations
type ReplicationManager struct {
	mu sync.RWMutex

	// Replication state
	role              string // "master" or "slave"
	replicationID     string
	replicationOffset int64

	// Master state (when role is "master")
	connectedReplicas []*ReplicaConnection

	// Replica state (when role is "slave")
	masterConfig *ReplicaConfig
}

type ReplicaConnection struct {
	Conn net.Conn
}

func NewReplicationManager() *ReplicationManager {
	// replication ID (hardcoded for these stages)
	replID := "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"

	return &ReplicationManager{
		role:              "master",
		replicationID:     replID,
		replicationOffset: 0,
		connectedReplicas: make([]*ReplicaConnection, 0),
	}
}

func (rm *ReplicationManager) SetReplicaMode(masterConfig *ReplicaConfig) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.role = "slave"
	rm.masterConfig = masterConfig
}

func (rm *ReplicationManager) GetRole() string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.role
}

func (rm *ReplicationManager) GetReplicationID() string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.replicationID
}

func (rm *ReplicationManager) GetReplicationOffset() int64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.replicationOffset
}

func (rm *ReplicationManager) GetConnectedReplicasCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return len(rm.connectedReplicas)
}

func (rm *ReplicationManager) AddReplica(conn net.Conn) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	replicaConn := &ReplicaConnection{
		Conn: conn,
	}

	rm.connectedReplicas = append(rm.connectedReplicas, replicaConn)
}

func (rm *ReplicationManager) RemoveReplica(conn net.Conn) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for i, replica := range rm.connectedReplicas {
		if replica.Conn == conn {
			rm.connectedReplicas = append(rm.connectedReplicas[:i], rm.connectedReplicas[i+1:]...)
			break
		}
	}
}

func (rm *ReplicationManager) GetReplicas() []*ReplicaConnection {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	replicas := make([]*ReplicaConnection, len(rm.connectedReplicas))
	copy(replicas, rm.connectedReplicas)
	return replicas
}
