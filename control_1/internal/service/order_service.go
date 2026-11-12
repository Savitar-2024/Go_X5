package service

import (
	"context"
	"strings"
	"time"

	"example.com/order-service/internal/order"
)

type OrderService struct {
	repository order.Repository
}

func NewOrderService(repository order.Repository) *OrderService {
	return &OrderService{repository: repository}
}

func (s *OrderService) Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error) {
	if len(input.Items) == 0 {
		return order.Order{}, order.ErrInvalidOrder
	}

	var total int64

	for _, item := range input.Items {
		if strings.TrimSpace(item.Name) == "" ||
			item.Quantity <= 0 ||
			item.Price <= 0 {
			return order.Order{}, order.ErrInvalidOrder
		}

		total += item.Price * int64(item.Quantity)
	}

	now := time.Now()

	newOrder := order.Order{
		Items:      append([]order.Item(nil), input.Items...),
		TotalPrice: total,
		Status:     order.StatusNew,
		CreatedAt:  now,
		History: []order.StatusChange{
			{
				Status:    order.StatusNew,
				ChangedAt: now,
			},
		},
	}

	return s.repository.Create(ctx, newOrder)
}

func (s *OrderService) GetByID(ctx context.Context, id int64) (order.Order, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *OrderService) List(ctx context.Context) ([]order.Order, error) {
	return s.repository.List(ctx)
}

func (s *OrderService) UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error) {
	switch status {
	case order.StatusNew, order.StatusProcessing,
		order.StatusCompleted, order.StatusCancelled:
	default:
		return order.Order{}, order.ErrInvalidStatus
	}

	current, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return order.Order{}, err
	}

	allowed := false

	switch current.Status {
	case order.StatusNew:
		allowed = status == order.StatusProcessing ||
			status == order.StatusCancelled

	case order.StatusProcessing:
		allowed = status == order.StatusCompleted ||
			status == order.StatusCancelled
	}

	if !allowed {
		return order.Order{}, order.ErrInvalidStatusTransition
	}

	current.Status = status

	current.History = append(current.History, order.StatusChange{
		Status:    status,
		ChangedAt: time.Now(),
	})

	return s.repository.Update(ctx, current)
}

func (s *OrderService) History(ctx context.Context, id int64) ([]order.StatusChange, error) {
	current, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := make([]order.StatusChange, len(current.History))
	copy(result, current.History)

	return result, nil
}
