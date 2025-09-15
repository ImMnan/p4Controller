#!/bin/sh

echo "Creating DIRs for:\n $ROOT_DIR, $CHECKPOINT_DIR, $VERSION_DIR"
mkdir -p $ROOT_DIR
mkdir -p $CHECKPOINT_DIR
mkdir -p $VERSION_DIR

if [ -z "$P4D_IP" ]; then
  echo "P4D_IP is not set. Using hostname"
  P4D_IP=$(hostname)
  
fi

p4d -V

echo "Starting p4d with IP: $P4D_IP and PORT: $P4D_PORT \n"

p4d -r $ROOT_DIR -L log -p $P4D_IP:$P4D_PORT -d


# Keep container running
while true; do sleep 60; done 