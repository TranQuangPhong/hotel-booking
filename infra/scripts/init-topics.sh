#!/bin/bash

# Kafka broker address
BROKER="localhost:9092"

# List of topics cần tạo
TOPICS=(
  "user-requests"
  "event-requests"
  "booking-requests"
  "payment-requests"
  "notification-requests"
)

# Tạo từng topic nếu chưa tồn tại
for topic in "${TOPICS[@]}"; do
  echo "Creating topic: $topic"
  docker exec -it kafka kafka-topics \
    --create \
    --if-not-exists \
    --bootstrap-server $BROKER \
    --replication-factor 1 \
    --partitions 1 \
    --topic $topic
done

echo "Kafka topics initialized."
