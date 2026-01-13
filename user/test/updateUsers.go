package main

//
//import (
//	"context"
//	"time"
//
//	log "github.com/sirupsen/logrus"
//	"google.golang.org/grpc"
//
//	pb "user/api/server/userpublicapi"
//)
//
//func main() {
//	conn, err := grpc.Dial("localhost:8081", grpc.WithInsecure(), grpc.WithBlock())
//	if err != nil {
//		log.Fatalf("Не удалось подключиться: %v", err)
//	}
//	defer conn.Close()
//
//	client := pb.NewUserPublicAPIClient(conn)
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//
//	userID1 := "019bb3d1-ce50-7dac-a057-2ca737e72ef4" // Заменить на реальный ID первого пользователя
//
//	// Обновляем первого пользователя (заполняем все поля)
//	email1 := "111@example.com"
//	telegram1 := "@use9999999931"
//	statusActive := pb.UserStatus(1)
//	_, err = client.UpdateUser(ctx, &pb.UpdateUserRequest{
//		UserID:   userID1,
//		Email:    &email1,
//		Telegram: &telegram1,
//		Status:   statusActive,
//	})
//	if err != nil {
//		log.Fatalf("Ошибка обновления первого пользователя: %v", err)
//	}
//	log.Println("Первый пользователь обновлён: добавлены email, telegram, статус Active")

//_, err = client.BlockUser(ctx, &pb.BlockUserRequest{
//	UserID: userID1,
//})
//if err != nil {
//	log.Fatalf("Ошибка блокировки первого пользователя: %v", err)
//}
//log.Println("Первый пользователь заблокирован")
//
//// Удаляем поля у второго пользователя (оставляем только логин)
//var emptyEmail *string = nil
//var emptyTelegram *string = nil
//statusActive = pb.UserStatus(1)
//userID2 := "019bb3d1-ce59-72a6-8663-3a14690f6ff8"
//_, err = client.UpdateUser(ctx, &pb.UpdateUserRequest{
//	UserID:   userID2,
//	Email:    emptyEmail,
//	Telegram: emptyTelegram,
//	Status:   statusActive,
//})
//if err != nil {
//	log.Fatalf("Ошибка очистки полей второго пользователя: %v", err)
//}
//log.Println("У второго пользователя удалены email и telegram, установлен статус Blocked")
//}
