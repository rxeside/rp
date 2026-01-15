package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "user/api/server/userpublicapi"
)

func main() {
	// 1. Подключение к User Service
	conn, err := grpc.NewClient("localhost:8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewUserPublicAPIClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 2. Генерация уникального логина
	login := fmt.Sprintf("demo_user_%d", time.Now().Unix())
	email := fmt.Sprintf("%s@example.com", login)
	tg := fmt.Sprintf("@%s", login)

	fmt.Println("--- [CHOREOGRAPHY DEMO] Creating User ---")
	fmt.Printf("Request: Login=%s, Email=%s\n", login, email)

	// 3. Вызов gRPC CreateUser
	resp, err := client.CreateUser(ctx, &pb.CreateUserRequest{
		Login:    login,
		Email:    &email,
		Telegram: &tg,
	})
	if err != nil {
		log.Fatalf("Error creating user: %v", err)
	}

	fmt.Printf("✅ User Created with ID: %s\n", resp.GetUserID())
	fmt.Println("Check RabbitMQ and Temporal UI for 'CreateWalletWorkflow' and 'Notification'")
	fmt.Println("-----------------------------------------")
}
