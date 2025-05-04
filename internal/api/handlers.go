// internal/api/handlers.go
package api

import (
	"encoding/json"
	"net/http"

	"github.com/Nishithcs/bank-info/internal/domain"
	"github.com/Nishithcs/bank-info/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

// Handler contains all the HTTP handlers
type Handler struct {
	accountService service.AccountService
	validator      *validator.Validate
}

// NewHandler creates a new handler
func NewHandler(accountService service.AccountService) *Handler {
	return &Handler{
		accountService: accountService,
		validator:      validator.New(),
	}
}

// RegisterRoutes registers all routes
func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/accounts", h.CreateAccount).Methods("POST")
	router.HandleFunc("/api/accounts/{id}", h.GetAccount).Methods("GET")
	router.HandleFunc("/api/accounts/{id}/transactions", h.GetTransactionHistory).Methods("GET")
	router.HandleFunc("/api/transactions", h.ProcessTransaction).Methods("POST")
}

// CreateAccount handles account creation requests
func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.accountService.CreateAccount(r.Context(), req); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusAccepted, map[string]string{
		"status": "account creation in progress",
	})
}

// GetAccount handles account retrieval requests
func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	account, err := h.accountService.GetAccount(r.Context(), id)
	if err != nil {
		if err == domain.ErrAccountNotFound {
			respondWithError(w, http.StatusNotFound, "Account not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, account)
}

// ProcessTransaction handles deposit and withdrawal requests
func (h *Handler) ProcessTransaction(w http.ResponseWriter, r *http.Request) {
	var req domain.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.accountService.ProcessTransaction(r.Context(), req)
	if err != nil {
		if err == domain.ErrAccountNotFound {
			respondWithError(w, http.StatusNotFound, "Account not found")
			return
		}
		if err == domain.ErrInsufficientBalance {
			respondWithError(w, http.StatusBadRequest, "Insufficient balance")
			return
		}
		if err == domain.ErrInvalidAmount {
			respondWithError(w, http.StatusBadRequest, "Invalid amount")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusAccepted, response)
}

// GetTransactionHistory handles transaction history requests
func (h *Handler) GetTransactionHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	history, err := h.accountService.GetTransactionHistory(r.Context(), id)
	if err != nil {
		if err == domain.ErrAccountNotFound {
			respondWithError(w, http.StatusNotFound, "Account not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, history)
}

// Helper functions for HTTP responses
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Internal server error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}