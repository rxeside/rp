#!/bin/bash

BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}[CLEANUP]${NC} Deleting Kind cluster..."
kind delete cluster --name rp-practice 2>/dev/null

echo -e "${BLUE}[CLEANUP]${NC} Removing data directories (sudo)..."
sudo rm -rf data/master
sudo rm -rf data/worker1
sudo rm -rf data/worker2
sudo rm -rf data/worker3

echo -e "${BLUE}[CLEANUP]${NC} Removing generated binaries..."
rm -f user/bin/user
rm -f payment/bin/payment
rm -f notification/bin/notification
rm -f order/bin/order
rm -f product/bin/product

echo -e "${BLUE}[CLEANUP]${NC} Done. Environment is clean."