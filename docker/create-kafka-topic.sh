#!/bin/bash

# Wait for Kafka to be ready with retries
echo "Waiting for Kafka to be ready... 3"
max_retries=10
retry_count=0
sleep_time=5

until kafka-topics --bootstrap-server localhost:9092 --list &>/dev/null; do
  if [ $retry_count -ge $max_retries ]; then
    echo "Kafka is not ready after $max_retries attempts, exiting."
    exit 1
  fi

  echo "Kafka is not ready yet, retrying in $sleep_time seconds..."
  sleep $sleep_time
  retry_count=$((retry_count + 1))
  sleep_time=$((sleep_time * 2))  # Exponential backoff
done

# Create the topic
echo "Creating Kafka topic..."
kafka-topics --create --topic messages --partitions 1 --replication-factor 1 --if-not-exists --bootstrap-server localhost:9092