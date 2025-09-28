#!/bin/sh

if [ -z "$CTR_IP" ]; then
  echo "CTR_IP is not set. Using hostname"
  CTR_IP=$(hostname)
fi

P4PORT=${CTR_IP}:${CTR_PORT}

while [ "$SERVER_INIT" != "true" ]; do
  echo "Waiting for p4Controller to initialize the server..."
  sleep 30
done

if [ "$SERVER_TYPE" = "master" ]; then
  echo "Starting p4d commit-server with: $P4PORT"
else
  echo "Starting replica server with: $P4PORT"
fi

if [ -z "$RUN_COMMAND" ]; then
  echo "RUN_COMMAND is not set. Using default command"
  RUN_COMMAND="p4d -r \"$P4ROOT\" -p \"$P4PORT\" -d"
fi

echo "Running command: $RUN_COMMAND"
eval "$RUN_COMMAND"

# Keep the container running
while true; do
  sleep 3600
done