#!/bin/sh

if [ -z "$CTR_IP" ]; then
  echo "CTR_IP is not set. Using hostname"
  CTR_IP=$(hostname)
fi

P4PORT=${CTR_IP}:${P4C_PORT}

while [ "$P4C_INIT" != "true" ]; do
  echo "Waiting for p4Controller to initialize the server..."
  sleep 30
done

if [ "$P4C_SERVER_TYPE" = "master" ]; then
  echo "Starting p4d commit-server with: $P4PORT"
else
  echo "Starting replica server with: $P4PORT"
fi

if [ -z "$P4C_RUN_COMMAND" ]; then
  echo "RUN_COMMAND is not set. Using default command"
  P4C_RUN_COMMAND="p4d -r \"$P4ROOT\" -p \"$P4PORT\" -d"
fi

echo "Running command: $P4C_RUN_COMMAND"
eval "$P4C_RUN_COMMAND"

# Keep the container running
while true; do
  sleep 3600
done