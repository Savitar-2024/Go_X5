package memory

import (
	"context"
	"sync"

	"example.com/order-service/internal/order"
)

type OrderRepository struct {
	mu     sync.RWMutex
	items  []order.Order
	nextID int64
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		items:  make([]order.Order, 0),
		nextID: 1,
	}
}

func copyOrder(value order.Order) order.Order {
	result := value

	result.Items = append([]order.Item(nil), value.Items...)
	result.History = append([]order.StatusChange(nil), value.History...)

	return result
}

func (r *OrderRepository) Create(
	_ context.Context,
	value order.Order,
) (order.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	value.ID = r.nextID
	r.nextID++

	r.items = append(r.items, copyOrder(value))

	return copyOrder(value), nil
}

func (r *OrderRepository) GetByID(
	_ context.Context,
	id int64,
) (order.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, item := range r.items {
		if item.ID == id {
			return copyOrder(item), nil
		}
	}

	return order.Order{}, order.ErrOrderNotFound
}

func (r *OrderRepository) List(
	_ context.Context,
) ([]order.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]order.Order, len(r.items))

	for i, item := range r.items {
		result[i] = copyOrder(item)
	}

	return result, nil
}

func (r *OrderRepository) Update(
	_ context.Context,
	value order.Order,
) (order.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, item := range r.items {
		if item.ID == value.ID {
			r.items[i] = copyOrder(value)
			return copyOrder(value), nil
		}
	}

	return order.Order{}, order.ErrOrderNotFound
}
