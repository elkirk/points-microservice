// This package defines the core data structures used in the app.
// PriorityQueue is a modified implementation of Google's example
// https://pkg.go.dev/container/heap
package models

import (
	"container/heap"
	"time"
)

type Payer struct {
	Payer string `json:"payer"`
}

// A Transaction is something we manage in a priority queue.
// The index is needed by heap.update and is maintained by
// the heap.Interface methods. We define a custom Less method
// to use Timestamp as the priority field
type Transaction struct {
	Payer
	Points    int       `json:"points"`
	Timestamp time.Time `json:"timestamp"`
	index     int
}

// A PriorityQueue implements heap.Interface and holds Transactions.
type PriorityQueue []*Transaction

// A TransactionStore will store all transactions and only grows.
type TransactionStore map[string][]*Transaction

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the oldest Transaction, so
	return pq[i].Timestamp.Before(pq[j].Timestamp)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*Transaction)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(transaction *Transaction,
	payer Payer, points int, timestamp time.Time) {
	transaction.Payer = payer
	transaction.Points = points
	transaction.Timestamp = timestamp
	heap.Fix(pq, transaction.index)
}
