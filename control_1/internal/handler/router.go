package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"example.com/order-service/internal/order"
	"github.com/go-chi/chi/v5"
)

type OrderService interface {
	Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error)
	GetByID(ctx context.Context, id int64) (order.Order, error)
	List(ctx context.Context) ([]order.Order, error)
	UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error)
	History(ctx context.Context, id int64) ([]order.StatusChange, error)
}

type Handler struct {
	service OrderService
	logger  *slog.Logger
}

type UpdateStatusRequest struct {
	Status order.Status `json:"status" example:"processing"`
}

func NewRouter(service OrderService, logger *slog.Logger) http.Handler {
	h := &Handler{service: service, logger: logger}

	router := chi.NewRouter()
	router.Use(requestLoggingMiddleware(logger))
	router.Use(recoveryMiddleware(logger))

	router.Get("/health", h.health)

	router.Route("/api/v1/orders", func(r chi.Router) {
		r.Post("/", h.createOrder)
		r.Get("/", h.listOrders)
		r.Get("/{id}", h.getOrder)
		r.Patch("/{id}/status", h.updateStatus)
		r.Get("/{id}/history", h.getHistory)
	})

	return router
}

// health godoc
// @Summary Проверка работоспособности
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// createOrder godoc
// @Summary Создать заказ
// @Description Создаёт заказ с начальным статусом new.
// @Tags Orders
// @Accept json
// @Produce json
// @Param order body order.CreateOrderInput true "Данные заказа"
// @Success 201 {object} order.Order
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders [post]
func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var input order.CreateOrderInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid JSON body")
		return
	}

	created, err := h.service.Create(r.Context(), input)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

// listOrders godoc
// @Summary Получить список заказов
// @Tags Orders
// @Produce json
// @Success 200 {array} order.Order
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders [get]
func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.service.List(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, orders)
}

// getOrder godoc
// @Summary Получить заказ по ID
// @Tags Orders
// @Produce json
// @Param id path int true "ID заказа"
// @Success 200 {object} order.Order
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders/{id} [get]
func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	result, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// updateStatus godoc
// @Summary Изменить статус заказа
// @Description Допустимые переходы: new в processing или cancelled; processing в completed или cancelled.
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "ID заказа"
// @Param request body UpdateStatusRequest true "Новый статус"
// @Success 200 {object} order.Order
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 409 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders/{id}/status [patch]
func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var input UpdateStatusRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid JSON body")
		return
	}

	result, err := h.service.UpdateStatus(r.Context(), id, input.Status)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// getHistory godoc
// @Summary Получить историю статусов
// @Tags Orders
// @Produce json
// @Param id path int true "ID заказа"
// @Success 200 {array} order.StatusChange
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders/{id}/history [get]
func (h *Handler) getHistory(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	history, err := h.service.History(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, history)
}

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid order ID")
		return 0, false
	}

	return id, true
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, order.ErrOrderNotFound):
		writeError(w, http.StatusNotFound, "ORDER_NOT_FOUND", err.Error())

	case errors.Is(err, order.ErrInvalidOrder):
		writeError(w, http.StatusBadRequest, "INVALID_ORDER", err.Error())

	case errors.Is(err, order.ErrInvalidStatus):
		writeError(w, http.StatusBadRequest, "INVALID_STATUS", err.Error())

	case errors.Is(err, order.ErrInvalidStatusTransition):
		writeError(w, http.StatusConflict, "INVALID_STATUS_TRANSITION", err.Error())

	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{
		Error: message,
		Code:  code,
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
