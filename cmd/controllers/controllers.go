package controllers

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/elkirk/points/cmd/models"
	"github.com/go-chi/chi/v5"
)

type Controller struct {
	models.TransactionStore // map[string][]*Transaction
	models.PriorityQueue    // []*Transaction
}

func (ctrl *Controller) Store(t models.Transaction) {
	payer := strings.ToUpper(t.Payer.Payer)
	ctrl.TransactionStore[payer] = append(ctrl.TransactionStore[payer], &t)
}

func (ctrl *Controller) AddHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var t models.Transaction
	err := decoder.Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	ctrl.Store(t)
	heap.Push(&ctrl.PriorityQueue, &t)
	// item := heap.Pop(&ctrl.PriorityQueue).(*models.Transaction)
	fmt.Printf("Length of queue: %+v\n", ctrl.PriorityQueue.Len())

}

func (ctrl *Controller) CheckEnoughPoints(spendAmount int) bool {
	payerBalances := map[string]int{}
	for _, payer := range ctrl.TransactionStore {
		for _, transaction := range payer {
			payerBalances[transaction.Payer.Payer] += transaction.Points
		}
	}
	var totalPoints int = 0
	for _, val := range payerBalances {
		totalPoints += val
	}
	return totalPoints >= spendAmount
}

type SpendRequest struct {
	Points int `json:"points"`
}

func (ctrl *Controller) SpendHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var req SpendRequest
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !ctrl.CheckEnoughPoints(req.Points) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Not enough points to cover spend request"}`))
		return
	}
	log.Printf("Received POST: %+v\n", req)

	spendFulfiller := map[string]int{}
	spentPoints := 0

	for spentPoints < req.Points {
		item := heap.Pop(&ctrl.PriorityQueue)
		transaction, ok := item.(models.Transaction)
		if !ok {
			log.Printf("Item popped from queue doesn't look like a transaction: %v\n", item)
		}

		if transaction.Points <= (req.Points - spentPoints) {
			spendFulfiller[transaction.Payer.Payer] -= transaction.Points

			// Transaction is removed from queue with Pop.
			// New Transaction with inverse points value is
			// added to Controller.Payers
			now := time.Now()
			t := models.Transaction{
				Payer:     transaction.Payer,
				Points:    transaction.Points,
				Timestamp: now,
			}
			ctrl.Store(t)

			spentPoints += transaction.Points

			// If current transaction can fullfill spend request, add a negative
			// points transaction to Controller.Payers, and push an updated
			// Transaction to Controller.Queue, preserving the original timestamp
		} else if transaction.Points > (req.Points - spentPoints) {
			remainder := req.Points - spentPoints
			spendFulfiller[transaction.Payer.Payer] -= remainder

			now := time.Now()

			t1 := &models.Transaction{
				Payer:     transaction.Payer,
				Points:    (transaction.Points - remainder),
				Timestamp: transaction.Timestamp,
			}
			heap.Push(&ctrl.PriorityQueue, t1)

			t2 := models.Transaction{
				Payer:     transaction.Payer,
				Points:    -remainder,
				Timestamp: now,
			}
			ctrl.Store(t2)

			spentPoints += remainder
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(spendFulfiller)
}

// checkStore returns the contents of TransactionsByPayer
func (ctrl *Controller) CheckStore(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ctrl.TransactionStore)
}

// checkQueue returns the contents of the queue
func (ctrl *Controller) CheckQueue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ctrl.PriorityQueue)
}

// checkQueue returns the contents of the queue
func (ctrl *Controller) DrainQueue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	queueContents := make([]*models.Transaction, 0)
	for ctrl.PriorityQueue.Len() > 0 {
		transaction := heap.Pop(&ctrl.PriorityQueue).(*models.Transaction)
		fmt.Printf("Transaction popped: %+v\n", transaction)
		queueContents = append(queueContents, transaction)

	}
	json.NewEncoder(w).Encode(queueContents)
}

// GET balance/{payer}
func (ctrl *Controller) BalanceByPayer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	payer := strings.ToUpper(chi.URLParam(r, "payer"))
	payerTransactions := ctrl.TransactionStore[payer]
	payerBalance := map[string]int{}
	for _, t := range payerTransactions {
		payerBalance[t.Payer.Payer] += t.Points
	}
	json.NewEncoder(w).Encode(payerBalance)
}

// GET /balance
func (ctrl *Controller) BalanceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	payerBalances := map[string]int{}
	for _, payer := range ctrl.TransactionStore {
		for _, transaction := range payer {
			payerBalances[transaction.Payer.Payer] += transaction.Points

		}
	}
	json.NewEncoder(w).Encode(payerBalances)
}
