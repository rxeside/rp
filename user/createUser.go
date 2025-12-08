package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	pb "user/api/server/userpublicapi"
)

func main() {
	conn, err := grpc.Dial("localhost:8081", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Не удалось подключиться: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserPublicAPIClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.StoreUser(ctx, &pb.StoreUserRequest{
		Login: "testuser-8",
	})
	if err != nil {
		log.Fatalf("Ошибка вызова: %v", err)
	}

	log.Printf("Создан пользователь: %s", resp.GetUserID())
}
