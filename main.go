package main

import (
	"container/heap"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/elkirk/points/cmd/controllers"
	"github.com/elkirk/points/cmd/models"
)

func main() {
	r := chi.NewRouter()

	// Setup in-memory storage
	payers := make(models.TransactionStore, 0)
	queue := make(models.PriorityQueue, 0)

	// Set up controller
	Controller := controllers.Controller{
		payers,
		queue,
	}

	heap.Init(&Controller.PriorityQueue)

	r.Use(middleware.Logger)

	r.Post("/add-transaction", Controller.AddHandler)
	r.Post("/spend", Controller.SpendHandler)
	r.Get("/check", Controller.CheckStore)
	r.Get("/queue", Controller.CheckQueue)
	r.Get("/queue/drain", Controller.DrainQueue)
	r.Get("/balance", Controller.BalanceHandler)
	r.Get("/balance/{payer}", Controller.BalanceByPayer)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	fmt.Println("Starting server on :3000")
	http.ListenAndServe(":3000", r)
}
