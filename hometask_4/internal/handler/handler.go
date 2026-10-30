package handler

import "example.com/orders/internal/model"

type Creator interface {
    Create(customer string) (model.Order, error)
}

type Handler struct {
    service Creator
}

func New(service Creator) *Handler {
    return &Handler{service: service}
}

func (h *Handler) Create(customer string) (model.Order, error) {
    return h.service.Create(customer)
}
