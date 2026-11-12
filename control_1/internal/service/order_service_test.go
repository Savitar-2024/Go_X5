package service_test

import (
	"context"
	"errors"
	"testing"

	"example.com/order-service/internal/order"
	"example.com/order-service/internal/repository/memory"
	"example.com/order-service/internal/service"
)

func newService() *service.OrderService {
	return service.NewOrderService(memory.NewOrderRepository())
}

func validInput() order.CreateOrderInput {
	return order.CreateOrderInput{
		Items: []order.Item{
			{Name: "Keyboard", Quantity: 2, Price: 500},
			{Name: "Mouse", Quantity: 1, Price: 300},
		},
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name    string
		input   order.CreateOrderInput
		wantErr error
	}{
		{
			name:  "valid order",
			input: validInput(),
		},
		{
			name:    "empty items",
			input:   order.CreateOrderInput{},
			wantErr: order.ErrInvalidOrder,
		},
		{
			name: "empty name",
			input: order.CreateOrderInput{Items: []order.Item{
				{Name: " ", Quantity: 1, Price: 100},
			}},
			wantErr: order.ErrInvalidOrder,
		},
		{
			name: "invalid quantity",
			input: order.CreateOrderInput{Items: []order.Item{
				{Name: "Book", Quantity: 0, Price: 100},
			}},
			wantErr: order.ErrInvalidOrder,
		},
		{
			name: "invalid price",
			input: order.CreateOrderInput{Items: []order.Item{
				{Name: "Book", Quantity: 1, Price: -100},
			}},
			wantErr: order.ErrInvalidOrder,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newService()

			got, err := s.Create(context.Background(), tt.input)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got error %v, want %v", err, tt.wantErr)
			}

			if tt.wantErr != nil {
				return
			}

			if got.ID != 1 {
				t.Errorf("ID = %d, want 1", got.ID)
			}
			if got.TotalPrice != 1300 {
				t.Errorf("total = %d, want 1300", got.TotalPrice)
			}
			if got.Status != order.StatusNew {
				t.Errorf("status = %s, want new", got.Status)
			}
			if len(got.History) != 1 {
				t.Errorf("history length = %d, want 1", len(got.History))
			}
		})
	}
}

func TestGetByIDAndList(t *testing.T) {
	ctx := context.Background()
	s := newService()

	first, err := s.Create(ctx, validInput())
	if err != nil {
		t.Fatal(err)
	}

	second, err := s.Create(ctx, validInput())
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr error
	}{
		{"first", first.ID, nil},
		{"second", second.ID, nil},
		{"not found", 999, order.ErrOrderNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.GetByID(ctx, tt.id)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got error %v, want %v", err, tt.wantErr)
			}
			if err == nil && got.ID != tt.id {
				t.Errorf("ID = %d, want %d", got.ID, tt.id)
			}
		})
	}

	orders, err := s.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(orders) != 2 {
		t.Errorf("list length = %d, want 2", len(orders))
	}
}

func TestUpdateStatus(t *testing.T) {
	tests := []struct {
		name    string
		from    []order.Status
		to      order.Status
		wantErr error
	}{
		{
			name: "new to processing",
			to:   order.StatusProcessing,
		},
		{
			name: "new to cancelled",
			to:   order.StatusCancelled,
		},
		{
			name: "processing to completed",
			from: []order.Status{order.StatusProcessing},
			to:   order.StatusCompleted,
		},
		{
			name: "processing to cancelled",
			from: []order.Status{order.StatusProcessing},
			to:   order.StatusCancelled,
		},
		{
			name:    "new to completed",
			to:      order.StatusCompleted,
			wantErr: order.ErrInvalidStatusTransition,
		},
		{
			name:    "unknown status",
			to:      order.Status("unknown"),
			wantErr: order.ErrInvalidStatus,
		},
		{
			name: "completed is final",
			from: []order.Status{
				order.StatusProcessing,
				order.StatusCompleted,
			},
			to:      order.StatusCancelled,
			wantErr: order.ErrInvalidStatusTransition,
		},
		{
			name:    "same status",
			to:      order.StatusNew,
			wantErr: order.ErrInvalidStatusTransition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := newService()

			created, err := s.Create(ctx, validInput())
			if err != nil {
				t.Fatal(err)
			}

			for _, status := range tt.from {
				if _, err := s.UpdateStatus(ctx, created.ID, status); err != nil {
					t.Fatal(err)
				}
			}

			got, err := s.UpdateStatus(ctx, created.ID, tt.to)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got error %v, want %v", err, tt.wantErr)
			}
			if err == nil && got.Status != tt.to {
				t.Errorf("status = %s, want %s", got.Status, tt.to)
			}
		})
	}
}

func TestHistory(t *testing.T) {
	ctx := context.Background()
	s := newService()

	created, err := s.Create(ctx, validInput())
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.UpdateStatus(ctx, created.ID, order.StatusProcessing)
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.UpdateStatus(ctx, created.ID, order.StatusCompleted)
	if err != nil {
		t.Fatal(err)
	}

	history, err := s.History(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}

	expected := []order.Status{
		order.StatusNew,
		order.StatusProcessing,
		order.StatusCompleted,
	}

	if len(history) != len(expected) {
		t.Fatalf("history length = %d, want %d", len(history), len(expected))
	}

	for i, entry := range history {
		if entry.Status != expected[i] {
			t.Errorf("history[%d] = %s, want %s", i, entry.Status, expected[i])
		}
	}

	_, err = s.History(ctx, 999)
	if !errors.Is(err, order.ErrOrderNotFound) {
		t.Errorf("got error %v, want ErrOrderNotFound", err)
	}
}
