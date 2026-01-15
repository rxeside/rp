#!/bin/bash

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

log() { echo -e "${BLUE}[SETUP]${NC} $1"; }
success() { echo -e "${GREEN}[READY]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }

# ==========================================
# 0. ОЧИСТКА (CLEANUP)
# ==========================================
log "=== PHASE 0: CLEANUP ==="
echo "Deleting Kind cluster..."
kind delete cluster --name rp-practice 2>/dev/null || true

echo "Removing data directories (sudo required)..."
sudo rm -rf data/master
sudo rm -rf data/worker1
sudo rm -rf data/worker2
sudo rm -rf data/worker3

echo "Removing generated binaries..."
rm -f user/bin/user
rm -f payment/bin/payment
rm -f notification/bin/notification
rm -f order/bin/order
rm -f product/bin/product

# ==========================================
# 1. СБОРКА (BUILD)
# ==========================================
log "=== PHASE 1: BUILD & CLUSTER CREATION ==="
./scripts/create-kind-cluster.sh

SERVICES=("user" "payment" "notification" "order" "product")

for svc in "${SERVICES[@]}"; do
    echo -n "Processing $svc... "
    if [ -d "$svc" ]; then
        # Компиляция Go
        cd "$svc"
        CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/$svc ./cmd/$svc
        cd ..

        # Сборка Docker и загрузка в Kind
        docker build -t "$svc:latest" ./$svc > /dev/null 2>&1
        kind load docker-image "$svc:latest" --name rp-practice
        echo "Done."
    else
        error "$svc not found!"
        exit 1
    fi
done

# ==========================================
# 2. ИНФРАСТРУКТУРА (INFRASTRUCTURE)
# ==========================================
log "=== PHASE 2: DEPLOYING INFRASTRUCTURE ==="
# Сюда входит mysql, rabbitmq, temporal и monitoring (prometheus/grafana)
kubectl apply -k config/infrastructure

log "Waiting for MySQL to be ready..."
kubectl wait -n infrastructure --for=condition=ready pod -l app=mysql --timeout=300s

log "Initializing Databases..."
sleep 5
# Создание БД и пользователей
kubectl exec -n infrastructure mysql-0 -- mysql -uroot -prootpass -e "
CREATE DATABASE IF NOT EXISTS user_db;
CREATE DATABASE IF NOT EXISTS payment_db;
CREATE DATABASE IF NOT EXISTS notification_db;
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

GRANT ALL PRIVILEGES ON *.* TO 'root'@'%';
GRANT ALL PRIVILEGES ON user_db.* TO 'user'@'%';
GRANT ALL PRIVILEGES ON payment_db.* TO 'payment'@'%';
GRANT ALL PRIVILEGES ON notification_db.* TO 'notification'@'%';
GRANT ALL PRIVILEGES ON order_db.* TO 'order'@'%';
GRANT ALL PRIVILEGES ON product.* TO 'product'@'%';
GRANT ALL PRIVILEGES ON temporal.* TO 'temporal'@'%';
GRANT ALL PRIVILEGES ON temporal_visibility.* TO 'temporal'@'%';
FLUSH PRIVILEGES;
" 2>/dev/null

log "Waiting for Dependencies (RabbitMQ, Temporal)..."
kubectl wait -n infrastructure --for=condition=ready pod -l app=rabbitmq --timeout=180s
# Перезапуск Temporal, чтобы он подцепил созданную БД
kubectl rollout restart deployment/temporal -n infrastructure > /dev/null 2>&1
kubectl wait -n infrastructure --for=condition=ready pod -l app=temporal --timeout=300s

# ==========================================
# 3. ПРИЛОЖЕНИЯ (APPLICATIONS)
# ==========================================
log "=== PHASE 3: DEPLOYING APPS ==="
kubectl apply -k config/application

log "Waiting for Applications to start..."
sleep 5
kubectl wait -n application --for=condition=ready pod --all --timeout=500s

log "⏳ Waiting 20s for DB Migrations to finish..."
sleep 20

# ==========================================
# 4. ДАННЫЕ (SEED DATA)
# ==========================================
log "=== PHASE 4: SEEDING DATA ==="
kubectl exec -n infrastructure mysql-0 -- mysql -uroot -prootpass -e "
USE user_db;
INSERT IGNORE INTO user (user_id, login, status, email, created_at, updated_at) VALUES ('22222222-2222-2222-2222-222222222222', 'test_user', 1, 'test@demo.com', NOW(), NOW());
USE payment_db;
INSERT IGNORE INTO wallet (wallet_id, user_id, balance, created_at, updated_at) VALUES (UUID(), '22222222-2222-2222-2222-222222222222', 1000.00, NOW(), NOW());
USE product;
INSERT IGNORE INTO products (id, name, price, quantity, created_at, updated_at) VALUES ('11111111-1111-1111-1111-111111111111', 'iPhone 16', 100.00, 10, NOW(), NOW());
" 2>/dev/null || true

# ==========================================
# 5. ПРОБРОС ПОРТОВ (PORT FORWARDING)
# ==========================================
log "=== PHASE 5: PORT FORWARDING ==="
pkill -f "kubectl port-forward" || true

# Приложения
kubectl port-forward svc/user -n application 8081:8081 > /dev/null 2>&1 &
PID_USER=$!
# Порт для метрик User Service (если вы хотите смотреть их curl-ом локально)
kubectl port-forward svc/user -n application 8082:8082 > /dev/null 2>&1 &
PID_USER_METRICS=$!

kubectl port-forward svc/order -n application 8084:8081 > /dev/null 2>&1 &
PID_ORDER=$!

# Инфраструктура
kubectl port-forward svc/temporal-ui -n infrastructure 8080:8080 > /dev/null 2>&1 &
PID_TEMPORAL=$!

# Мониторинг
kubectl port-forward svc/grafana -n infrastructure 3000:3000 > /dev/null 2>&1 &
PID_GRAFANA=$!
kubectl port-forward svc/prometheus -n infrastructure 9090:9090 > /dev/null 2>&1 &
PID_PROMETHEUS=$!

success "System is UP and RUNNING!"
echo "------------------------------------------------"
echo -e "User Service (gRPC):  localhost:8081"
echo -e "User Metrics:         http://localhost:8082/metrics"
echo -e "Temporal UI:          http://localhost:8080"
echo -e "Grafana:              http://localhost:3000 (admin/admin)"
echo -e "Prometheus:           http://localhost:9090"
echo "------------------------------------------------"
echo "Press Ctrl+C to stop port forwarding and exit."

trap "kill $PID_USER $PID_USER_METRICS $PID_ORDER $PID_TEMPORAL $PID_GRAFANA $PID_PROMETHEUS 2>/dev/null; exit" INT
wait