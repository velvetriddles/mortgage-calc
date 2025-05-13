package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/velvetriddles/mortgage-calc/internal/cache"
	"github.com/velvetriddles/mortgage-calc/internal/model"
	"github.com/velvetriddles/mortgage-calc/internal/service"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Result model.ExecuteResponse `json:"result"`
}

func writeJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func writeErrorResponse(w http.ResponseWriter, message string, status int) {
	writeJSON(w, ErrorResponse{Error: message}, status)
}

type MortHandler struct {
	cache      *cache.MortCache
	calculator service.Calculator
}

func NewMortHandler(cache *cache.MortCache, calculator service.Calculator) *MortHandler {
	return &MortHandler{
		cache:      cache,
		calculator: calculator,
	}
}

func validateProgramRequest(program model.ProgramRequest) error {
	count := 0
	if program.Salary {
		count++
	}
	if program.Military {
		count++
	}
	if program.Base {
		count++
	}

	switch {
	case count == 0:
		return model.ErrChooseNone
	case count > 1:
		return model.ErrChooseMultiple
	}

	return nil
}

func getErrorMessage(err error) string {
	switch err {
	case model.ErrChooseNone:
		return "choose program"
	case model.ErrChooseMultiple:
		return "choose only 1 program"
	case model.ErrInitialPaymentLow:
		return "the initial payment should be more"
	case service.ErrNoProgramSelected:
		return "no mortgage program selected"
	default:
		return err.Error()
	}
}

func (h *MortHandler) Execute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, "Method not supported", http.StatusMethodNotAllowed)
		return
	}

	var req model.ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := validateProgramRequest(req.Program); err != nil {
		writeErrorResponse(w, getErrorMessage(err), http.StatusBadRequest)
		return
	}

	agg, err := h.calculator.Calculate(req, time.Now())
	if err != nil {
		writeErrorResponse(w, getErrorMessage(err), http.StatusBadRequest)
		return
	}

	resp := model.ExecuteResponse{
		Params: model.RequestParams{
			ObjectCost:     req.ObjectCost,
			InitialPayment: req.InitialPayment,
			Months:         req.Months,
		},
		Program:    req.Program,
		Aggregates: agg,
	}

	h.cache.Save(resp)
	writeJSON(w, SuccessResponse{Result: resp}, http.StatusOK)
}

func (h *MortHandler) GetCache(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, "Method not supported", http.StatusMethodNotAllowed)
		return
	}

	items, err := h.cache.GetAll()
	if err != nil || len(items) == 0 {
		writeErrorResponse(w, "empty cache", http.StatusBadRequest)
		return
	}

	writeJSON(w, items, http.StatusOK)
}
