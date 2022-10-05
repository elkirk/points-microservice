package main

import (
	"fmt"
	"time"
)

// Queue—Array of Transaction Type
type Queue []*Transaction

type Payer struct {
	Payer string `json:"payer"`
}

type Transaction struct {
	Payer
	Points    int       `json:"points"`
	Timestamp time.Time `json:"-"`
}

func main() {
	queue := make(Queue, 0)

	var transaction1 *Transaction = &Transaction{}
	var transaction2 *Transaction = &Transaction{}
	var transaction3 *Transaction = &Transaction{}
	var transaction4 *Transaction = &Transaction{}
	var transaction5 *Transaction = &Transaction{}

	transaction1.New("DANNON", 1000, "2020-11-02T14:00:00Z")
	transaction2.New("UNILEVER", 200, "2020-10-31T11:00:00Z")
	transaction3.New("DANNON", -200, "2020-10-31T15:00:00Z")
	transaction4.New("MILLER COORS", 10000, "2020-11-01T14:00:00Z")
	transaction5.New("DANNON", 300, "2020-10-31T10:00:00Z")

	queue.Add(transaction1)
	queue.Add(transaction2)
	queue.Add(transaction3)
	queue.Add(transaction4)
	queue.Add(transaction5)

	for i, _ := range queue {
		fmt.Println(queue[i])
	}
}

// New method initializes Transaction with payer, points, and timestamp
func (transaction *Transaction) New(payer string, points int, timestamp string) {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		fmt.Println(err)
	}
	transaction.Payer = Payer{payer}
	transaction.Points = points
	transaction.Timestamp = t
}

//Add method adds the transaction to the queue
func (queue *Queue) Add(transaction *Transaction) {
	if len(*queue) == 0 {
		*queue = append(*queue, transaction)
	} else {
		var appended bool
		appended = false
		var i int
		var queuedTransaction *Transaction
		for i, queuedTransaction = range *queue {
			if transaction.Timestamp.Before(queuedTransaction.Timestamp) {
				*queue = append((*queue)[:i], append(Queue{transaction}, (*queue)[i:]...)...)
				appended = true
				break
			}
		}
		if !appended {
			*queue = append(*queue, transaction)
		}
	}
}
