#!/bin/bash

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

log() {
    echo -e "${BLUE}[DEMO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 1. Проверка окружения
command -v docker >/dev/null 2>&1 || { error "Docker required"; exit 1; }
command -v kind >/dev/null 2>&1 || { error "Kind required"; exit 1; }
command -v kubectl >/dev/null 2>&1 || { error "Kubectl required"; exit 1; }

# 2. Очистка старого кластера
if kind get clusters | grep -q "rp-practice"; then
    log "Cluster 'rp-practice' exists. Cleaning up..."
    ./scripts/delete-kind-cluster.sh > /dev/null 2>&1 || true
fi

# 3. Создание кластера
log "Creating Kubernetes cluster..."
./scripts/create-kind-cluster.sh

# 4. Сборка приложений
SERVICES=("user" "payment" "notification" "order" "product")
log "Building binaries and Docker images..."

for svc in "${SERVICES[@]}"; do
    echo -n "Processing $svc... "
    if [ -d "$svc" ]; then
        cd "$svc"
        CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/$svc ./cmd/$svc
        if [ $? -ne 0 ]; then
            echo ""; error "Build failed for $svc"; exit 1
        fi
        cd ..
        docker build -t "$svc:latest" ./$svc > /dev/null 2>&1
        kind load docker-image "$svc:latest" --name rp-practice
        echo "Done."
    else
        error "Directory $svc not found!"
        exit 1
    fi
done

# 5. Деплой Инфраструктуры
log "Deploying Infrastructure..."
kubectl apply -k config/infrastructure

log "Waiting for MySQL..."
kubectl wait -n infrastructure --for=condition=ready pod -l app=mysql --timeout=300s

# --- Создание БД ---
log "Creating Databases & Users..."
sleep 5
kubectl exec -n infrastructure mysql-0 -- mysql -uroot -prootpass -e "
CREATE DATABASE IF NOT EXISTS user_db;
CREATE DATABASE IF NOT EXISTS payment_db;
CREATE DATABASE IF NOT EXISTS order_db;
CREATE DATABASE IF NOT EXISTS product;
CREATE DATABASE IF NOT EXISTS temporal;
CREATE DATABASE IF NOT EXISTS temporal_visibility;
CREATE USER IF NOT EXISTS 'user'@'%' IDENTIFIED BY '12345Q';
CREATE USER IF NOT EXISTS 'payment'@'%' IDENTIFIED BY '12345Q';
CREATE USER IF NOT EXISTS 'order'@'%' IDENTIFIED BY '12345Q';
CREATE USER IF NOT EXISTS 'product'@'%' IDENTIFIED BY '12345Q';
CREATE USER IF NOT EXISTS 'notification'@'%' IDENTIFIED BY '12345Q';
CREATE USER IF NOT EXISTS 'temporal'@'%' IDENTIFIED BY '12345Q';
GRANT ALL PRIVILEGES ON user_db.* TO 'user'@'%';
GRANT ALL PRIVILEGES ON payment_db.* TO 'payment'@'%';
GRANT ALL PRIVILEGES ON payment_db.* TO 'notification'@'%';
GRANT ALL PRIVILEGES ON order_db.* TO 'order'@'%';
GRANT ALL PRIVILEGES ON product.* TO 'product'@'%';
GRANT ALL PRIVILEGES ON temporal.* TO 'temporal'@'%';
GRANT ALL PRIVILEGES ON temporal_visibility.* TO 'temporal'@'%';
FLUSH PRIVILEGES;
"

log "Waiting for RabbitMQ & Temporal..."
kubectl wait -n infrastructure --for=condition=ready pod -l app=rabbitmq --timeout=180s
kubectl rollout restart deployment/temporal -n infrastructure > /dev/null 2>&1
kubectl wait -n infrastructure --for=condition=ready pod -l app=temporal --timeout=300s

# 6. Деплой Приложений
log "Deploying Microservices..."
kubectl apply -k config/application

log "Waiting for Applications..."
sleep 5 # Даем время на создание подов
kubectl wait -n application --for=condition=ready pod --all --timeout=300s

# --- SEED DATA (Наполняем данными для теста) ---
log "Seeding Test Data (User, Wallet, Product)..."
# User ID: 2222... | Product ID: 1111... | Wallet Balance: 1000
kubectl exec -n infrastructure mysql-0 -- mysql -uroot -prootpass -e "
USE user_db;
INSERT INTO user (user_id, login, status, email, created_at, updated_at) VALUES ('22222222-2222-2222-2222-222222222222', 'test_user', 1, 'test@demo.com', NOW(), NOW());

USE payment_db;
INSERT INTO wallet (wallet_id, user_id, balance, created_at, updated_at) VALUES (UUID(), '22222222-2222-2222-2222-222222222222', 1000.00, NOW(), NOW());

USE product;
INSERT INTO products (id, name, price, quantity, created_at, updated_at) VALUES ('11111111-1111-1111-1111-111111111111', 'iPhone 16', 100.00, 10, NOW(), NOW());
"
success "Test Data Seeded!"
# -----------------------------------------------

# 7. Настройка Port-Forwarding
log "Setting up Port-Forwarding..."
pkill -f "kubectl port-forward" || true
kubectl port-forward svc/user -n application 8081:8081 > /dev/null 2>&1 &
PID_USER=$!
kubectl port-forward svc/payment -n application 8083:8081 > /dev/null 2>&1 &
PID_PAYMENT=$!
kubectl port-forward svc/order -n application 8084:8081 > /dev/null 2>&1 &
PID_ORDER=$!
kubectl port-forward svc/product -n application 8085:8081 > /dev/null 2>&1 &
PID_PRODUCT=$!
kubectl port-forward svc/temporal-ui -n infrastructure 8080:8080 > /dev/null 2>&1 &
PID_TEMPORAL=$!

sleep 5
#
## 8. Запуск Теста
#log "Running End-to-End Saga Test..."
#if [ ! -f "order/cmd/e2e/main.go" ]; then
#    error "Test file order/cmd/e2e/main.go not found!"
#    exit 1
#fi
#
## Запускаем тест из папки order, чтобы он видел go.mod
#cd order
#go run cmd/e2e/main.go
#TEST_EXIT_CODE=$?
#cd ..
#
## 9. Логи Notification
#echo ""
#log "Checking Notification Service logs:"
#echo "-----------------------------------------------------"
#NOTIF_POD=$(kubectl get pods -n application -l app=notification-handler -o jsonpath="{.items[0].metadata.name}")
#kubectl logs "$NOTIF_POD" -n application --tail=10
#echo "-----------------------------------------------------"
#
## 10. Очистка
#log "Stopping port-forwards..."
#kill $PID_USER $PID_PAYMENT $PID_ORDER $PID_PRODUCT $PID_TEMPORAL || true
#
#if [ $TEST_EXIT_CODE -eq 0 ]; then
#    echo ""
#    success "DEMO COMPLETED SUCCESSFULLY! 🎉"
#    echo "Temporal UI: http://localhost:8080"
#else
#    echo ""
#    error "DEMO FAILED."
#    exit $TEST_EXIT_CODE
#fi