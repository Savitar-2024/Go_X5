package main

import (
    "fmt"
    "log"

    "example.com/orders/internal/config"
    "example.com/orders/internal/handler"
    "example.com/orders/internal/repository"
    "example.com/orders/internal/service"
)

func main() {
    cfg := config.Load()

    repo := repository.NewMemory()
    svc := service.New(repo)
    h := handler.New(svc)

    order, err := h.Create("Покупатель")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf(
        "сервис слушал бы %s, создан заказ: %+v\n",
        cfg.HTTPAddr, order,
    )
}
