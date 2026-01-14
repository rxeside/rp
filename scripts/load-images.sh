#!/usr/bin/env bash
set -euo pipefail

# Сборка образов (предполагается, что brewkit build уже выполнен и бинарники есть)
# Для Kind лучше использовать локальные теги, чтобы он не пытался тянуть их из DockerHub

echo "Building Docker images..."
docker build -t user:latest ./user
docker build -t payment:latest ./payment
docker build -t notification:latest ./notification
docker build -t order:latest ./order
docker build -t product:latest ./product

echo "Loading images into Kind..."
kind load docker-image user:latest --name rp-practice
kind load docker-image payment:latest --name rp-practice
kind load docker-image notification:latest --name rp-practice
kind load docker-image order:latest --name rp-practice
kind load docker-image product:latest --name rp-practice