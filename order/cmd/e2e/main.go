package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "order/api/server/orderinternalapi"
)

func main() {
	// Подключаемся к Order Service (порт проброшен скриптом run_demo.sh)
	conn, err := grpc.NewClient("localhost:8084", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewOrderInternalAPIClient(conn)

	// Используем заранее подготовленные ID (которые мы вставим в БД через скрипт)
	userID := "22222222-2222-2222-2222-222222222222"
	productID := "11111111-1111-1111-1111-111111111111"

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("--- Sending StoreOrder Request (Starting SAGA) ---")
	// Создаем заказ
	resp, err := client.StoreOrder(ctx, &pb.StoreOrderRequest{
		CustomerID: userID,
		Status:     pb.OrderStatus_Open,
		Items: []*pb.OrderItem{
			{
				ProductID:  productID,
				Count:      1,
				TotalPrice: 100.0, // У нас на счету будет 1000, должно хватить
			},
		},
	})
	if err != nil {
		log.Fatalf("Error creating order: %v", err)
	}

	orderID := resp.OrderID
	fmt.Printf("Order Created with ID: %s\n", orderID)

	fmt.Println("--- Polling Order Status ---")
	// Опрашиваем статус, пока он не станет Paid или Cancelled
	for i := 0; i < 20; i++ {
		r, err := client.FindOrder(ctx, &pb.FindOrderRequest{OrderID: orderID})
		if err != nil {
			log.Printf("Getting order error: %v", err)
		} else {
			fmt.Printf("Attempt %d: Status = %s\n", i+1, r.Status)

			if r.Status == pb.OrderStatus_Paid {
				fmt.Println("🎉 SUCCESS: Order is PAID! Saga finished successfully.")
				return
			}
			if r.Status == pb.OrderStatus_Cancelled {
				log.Fatalf("❌ FAILURE: Order was CANCELLED. Saga compensation triggered (maybe insufficient funds/stock?).")
			}
		}
		time.Sleep(1 * time.Second)
	}
	log.Fatalf("❌ TIMEOUT: Order status stuck.")
}
