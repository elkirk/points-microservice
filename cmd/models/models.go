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

// This example creates a PriorityQueue with some items, adds and manipulates an item,
// and then removes the items in priority order.
// func Example_priorityQueue() {
// 	// Some items and their priorities.
// 	items := []string{
// 		`{ "payer": "DANNON", "points": 1000, "timestamp": "2020-11-02T14:00:00Z" }`,
// 		`{ "payer": "UNILEVER", "points": 200, "timestamp": "2020-10-31T11:00:00Z" }`,
// 		`{ "payer": "DANNON", "points": -200, "timestamp": "2020-10-31T15:00:00Z" }`,
// 		`{ "payer": "MILLER COORS", "points": 10000, "timestamp": "2020-11-01T14:00:00Z" }`,
// 		`{ "payer": "DANNON", "points": 300, "timestamp": "2020-10-31T10:00:00Z" } `,
// 	}
//
// 	// Create a priority queue, put the items in it, and
// 	// establish the priority queue (heap) invariants.
// 	pq := make(PriorityQueue, len(items))
// 	i := 0
// 	for item := range items {
// 		decoder := json.NewDecoder(item)
// 		var t Transaction
// 		err := decoder.Decode(&t)
// 		pq[i] = &Transaction{
// 			value:    value,
// 			priority: priority,
// 			index:    i,
// 		}
// 		i++
// 	}
// 	heap.Init(&pq)
//
// 	// Insert a new item and then modify its priority.
// 	item := &Item{
// 		value:    "orange",
// 		priority: 1,
// 	}
// 	heap.Push(&pq, item)
// 	pq.update(item, item.value, 5)
//
// 	// Take the items out; they arrive in decreasing priority order.
// 	for pq.Len() > 0 {
// 		item := heap.Pop(&pq).(*Item)
// 		fmt.Printf("%.2d:%s ", item.priority, item.value)
// 	}
// 	// Output:
// 	// 05:orange 04:pear 03:banana 02:apple
// }
