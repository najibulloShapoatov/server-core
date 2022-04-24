# Cluster

Allows communication and synchronization between cluster nodes

## Install

To install the library

```
$ go get github.com/najibulloShapoatov/server-core/cluster
```

## Usage example

```go

c, err := cluster.Join("cluster-name")

// Listen for cluster messages
c.OnMessage(func(msg *cluster.Message) {
   // handle message from other nodes
})

// send a message to all nodes in the cluster
c.Broadcast(message)

// Obtain a mutex lock in the cluster
err := c.Lock("lock-name")
if err != nil {
    // Mutex is already lock by another node
}

// ...
// Perform action
// ...

// release lock 
c.Unlock("lock-name")

// Leave the cluster
c.Leave()

```
