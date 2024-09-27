#!/bin/bash

echo "Starting script create kafka topic"

/bin/bash /tmp/create-kafka-topic.sh &

echo "Starting kafka server"

exec /etc/confluent/docker/run