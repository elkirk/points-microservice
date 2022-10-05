// This package contains handler functions for http routes defined in main.go
// Handler functions are methods that receive the Controller struct, so each handler
// will have access to the in-memory TransactionStore and PriorityQueue, which are
// defined in the models package.
package controllers

import (
	"container/heap"
	"encoding/json"
	"errors"
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

// Store takes a Transaction and stores it in the controller's TransactionStore
func (ctrl *Controller) Store(t models.Transaction) {
	payer := strings.ToUpper(t.Payer.Payer)
	ctrl.TransactionStore[payer] = append(ctrl.TransactionStore[payer], &t)
}

// AddHandler handles POST requests to the /add-transaction route.
// It first checks that the request body has the right format by decoding
// into a Transaction variable. If there is no error, it stores that Transaction
// in the Controller's TransactionStore and pushes the Transaction to the
// PriorityQueue
func (ctrl *Controller) AddHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var t models.Transaction
	err := decoder.Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if t.Points < 0 {
		err := ctrl.CheckNotNegative(t.Payer.Payer, t.Points)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`Transaction not added because it would cause Payer's points to go negative.`))
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	ctrl.Store(t)
	heap.Push(&ctrl.PriorityQueue, &t)

}

// CheckNotNegative is called in AddHandler when the to-be-added Transaction has
// a negative points balance. It checks TransactionStore to verify the negative
// Transaction will not result in a payer having a negative total points balance.
func (ctrl *Controller) CheckNotNegative(payer string, points int) error {
	payerBalances := map[string]int{}
	for _, payer := range ctrl.TransactionStore {
		for _, transaction := range payer {
			payerBalances[transaction.Payer.Payer] += transaction.Points
		}
	}
	if payerBalances[payer] < -points {
		return errors.New("Transaction would cause Payer's points to go negative")
	}
	return nil
}

// CheckEnoughPoints is called in SpendHandler. It checks TransactionStore
// and returns True if there are enough points to cover the SpendRequest,
// otherwise returns False
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

// SpendHandler handles POST requests to the /spend route. Spend requests are
// fulfilled using transactions in the PriorityQueue. Transactions are popped
// from the queue in order by timestamp, oldest first, and their points are
// added to the spentPoints variable. Loop continues iterating over Transactions
// in queue until spentPoints == SpendRequest.Points. When this condition is met,
// loop exits and an updated transaction is added to the queue, if the final
// Transaction has left-over points. Negative transactions representing expediture
// are added to TransactionStore, so that payer balances reflect the spend.
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

	spendFulfiller := map[string]int{}
	spentPoints := 0

	for spentPoints < req.Points {
		item := heap.Pop(&ctrl.PriorityQueue)
		transaction, ok := item.(*models.Transaction)
		if !ok {
			log.Printf("Item popped from queue doesn't look like a transaction: %v\n", item)
		}

		// If current transaction can't fulfill spend request, consume entire
		// transaction and add a new transaction with inverse points to TransactionStore
		if transaction.Points <= (req.Points - spentPoints) {
			spendFulfiller[transaction.Payer.Payer] -= transaction.Points

			now := time.Now()
			t := models.Transaction{
				Payer:     transaction.Payer,
				Points:    -transaction.Points,
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

			// t1 is the modified final transaction with preserved timestamp and updated
			// points value, which is re-added to the PriorityQueue
			t1 := &models.Transaction{
				Payer:     transaction.Payer,
				Points:    (transaction.Points - remainder),
				Timestamp: transaction.Timestamp,
			}
			heap.Push(&ctrl.PriorityQueue, t1)

			// t2 represents the negative transaction that fulfills the spend request
			// and is added to TransactionStore
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

// CheckStore returns the contents of TransactionStore, which are grouped by Payer
func (ctrl *Controller) CheckStore(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ctrl.TransactionStore)
}

// CheckQueue returns the contents of the queue, which will not be ordered.
func (ctrl *Controller) CheckQueue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ctrl.PriorityQueue)
}

// DrainQueue drains the queue in order of priority (oldest to newest).
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

// BalanceByPayer returns points balance for a specific payer
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

// BalanceHandler returns all payer point balances
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
