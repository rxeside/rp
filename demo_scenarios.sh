#!/bin/bash

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== SCENARIO 1: CHOREOGRAPHY (Event-Driven) ===${NC}"
echo "Действие: Создаем пользователя через gRPC."
echo "Ожидание: User Service сохраняет -> кидает Event -> Payment создает кошелек + Notification шлет письмо."
echo "---------------------------------------------------------"

cd user
go run cmd/e2e/main.go
cd ..

echo "1. Лог выше: 'User Created'"
echo "2. Проверка БД Payment (кошелек создан):"
kubectl exec -it -n infrastructure mysql-0 -- mysql -uroot -prootpass -e "SELECT * FROM payment_db.wallet ORDER BY created_at DESC LIMIT 1;" 2>/dev/null

echo "3. Проверка Метрик (Prometheus):"
# Получаем имя пода user-handler
HANDLER_POD=$(kubectl get pods -n application -l app=user-handler -o jsonpath="{.items[0].metadata.name}")
echo "   Под (Handler): $HANDLER_POD"

# Временно пробрасываем порт именно от этого пода в фон
kubectl port-forward "$HANDLER_POD" -n application 9999:8082 > /dev/null 2>&1 &
PF_PID=$!

# Даем пару секунд на поднятие туннеля
sleep 2

if curl -s http://localhost:9999/metrics | grep -q "app_events_processed_total"; then
    echo "   ✅ Метрика 'app_events_processed_total' найдена!"
    curl -s http://localhost:9999/metrics | grep "app_events_processed_total"
else
    echo "   ⚠️ Метрика пока не найдена (возможно, Prometheus еще не успел собрать или событие в пути)"
fi

# Убиваем временный проброс порта
kill $PF_PID 2>/dev/null
echo "4. Логи Notification Service (письмо отправлено):"
NOTIF_POD=$(kubectl get pods -n application -l app=notification-handler -o jsonpath="{.items[0].metadata.name}")
kubectl logs "$NOTIF_POD" -n application --tail=2

echo -e "\n---------------------------------------------------------"
read -p "Нажми Enter для запуска ОРКЕСТРАЦИИ (Saga)..."

echo -e "${BLUE}=== SCENARIO 2: ORCHESTRATION (Saga - Success) ===${NC}"
echo "Действие: Покупка iPhone (Цена: 100). Баланс: 1000."
echo "Ожидание: Reserve Product -> Charge Wallet -> Order Paid."
echo "---------------------------------------------------------"

cd order
go run cmd/e2e/main.go
cd ..

echo "1. Открой Temporal UI: http://localhost:8080"
echo "2. Покажи Workflow 'CreateOrderSaga' (Status: Completed)"
echo "3. В истории Activity: ReserveProduct (Done), ChargeWallet (Done)"

echo -e "\n---------------------------------------------------------"
read -p "Нажми Enter для запуска КОМПЕНСАЦИИ (Saga Rollback)..."

echo -e "${BLUE}=== SCENARIO 3: ORCHESTRATION (Saga - Compensation) ===${NC}"
echo "Действие: Покупка ОЧЕНЬ дорогого товара. Баланс: 900 (остаток)."
echo "Ожидание: Reserve Product (OK) -> Charge Wallet (FAIL) -> Release Product (Compensate) -> Order Cancelled."
echo "---------------------------------------------------------"

# Создаем временный тест для эмуляции ошибки
cat <<GO > order/cmd/e2e/fail_test.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"
    "github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "order/api/server/orderinternalapi"
)

func main() {
	conn, err := grpc.NewClient("localhost:8084", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil { log.Fatalf("Fail: %v", err) }
	defer conn.Close()
	client := pb.NewOrderInternalAPIClient(conn)

    // Тот же юзер, тот же продукт, но цена 1 000 000
	userID := "22222222-2222-2222-2222-222222222222"
	productID := "11111111-1111-1111-1111-111111111111"

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

    dummyID := uuid.New().String()

	fmt.Println("--- Покупаем товар за 999999.00 (Денег нет) ---")
	resp, err := client.StoreOrder(ctx, &pb.StoreOrderRequest{
		CustomerID: userID,
		Status:     pb.OrderStatus_Open,
		Items: []*pb.OrderItem{
            {
                OrderID:    dummyID, // Валидация требует UUID
                ProductID: productID,
                Count: 1,
                TotalPrice: 999999.00,
            },
        },
	})
	if err != nil { log.Fatalf("Error: %v", err) }

	fmt.Printf("Order Created ID: %s. Ждем отката саги...\n", resp.OrderID)

	for i := 0; i < 15; i++ {
		r, _ := client.FindOrder(ctx, &pb.FindOrderRequest{OrderID: resp.OrderID})
		if r.Status == pb.OrderStatus_Cancelled {
			fmt.Println("✅ SUCCESS: Order CANCELLED! Компенсация сработала.")
			return
		}
		time.Sleep(1 * time.Second)
	}
    fmt.Println("❌ Timeout waiting for cancellation.")
}
GO

cd order
go run cmd/e2e/fail_test.go
rm cmd/e2e/fail_test.go
cd ..

echo -e "\n${GREEN}Что показать преподавателю:${NC}"
echo "1. Temporal UI -> Новый Workflow."
echo "2. История: ReserveProduct (Green) -> ChargeWallet (Red/Fail) -> ReleaseProduct (Green/Compensate)."
echo "Это доказывает, что распределенная транзакция откатилась корректно."

echo -e "\n${BLUE}Демонстрация завершена!${NC}"