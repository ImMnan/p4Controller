#!/bin/sh

echo "Creating DIRs for:\n $ROOT_DIR, $CHECKPOINT_DIR, $VERSION_DIR"
mkdir -p $ROOT_DIR
mkdir -p $CHECKPOINT_DIR
mkdir -p $VERSION_DIR

p4d -V

p4d -r $ROOT_DIR -L log -p $P4D_IP:$P4D_PORT -d


# Keep container running
while true; do sleep 60; done 