#!/bin/sh

if [ -z "$CTR_IP" ]; then
  echo "CTR_IP is not set. Using hostname"
  CTR_IP=$(hostname)
fi

P4PORT=${CTR_IP}:${CTR_PORT}

if [ "$SERVER_TYPE" = "master" ]; then
  while [ "$SERVER_INIT" != "true" ]; do
    echo "Waiting for p4Controller to initialize the replica server..."
    sleep 30
  done
  echo "Starting p4d commit-server with: $P4PORT \n"
  if [ -z "$RUN_COMMAND" ]; then
    echo "RUN_COMMAND is not set. Using default command"
    RUN_COMMAND="p4d -r \"$P4ROOT\" -p \"$P4PORT\" -d"
  fi
  echo "Running command: $RUN_COMMAND"
  # Start p4d server by running the command in env $RUN_COMMAND
  eval "$RUN_COMMAND"
fi


if [ "$SERVER_TYPE" != "master" ]; then
  while [ "$SERVER_INIT" != "true" ]; do
    echo "Waiting for p4Controller to initialize the replica server..."
    sleep 30
  done
  echo "Starting replica server with: $P4PORT \n"
  if [ -z "$RUN_COMMAND" ]; then
    RUN_COMMAND="p4d -r \"$P4ROOT\" -p \"$P4PORT\" -d"
  fi
  echo "Running command: $RUN_COMMAND"
  eval "$RUN_COMMAND"
fi

# p4d -r $ROOT_DIR -L log --pid-file=/opt/p4d-root/p4d.pid -p $P4PORT -d

# Parent loop to check P4D_FLAG and control p4d
while true; do
  if [ "$P4D_FLAG" = "false" ]; then
    echo "P4D_FLAG is false, p4Controller will stop or has stopped the p4d server."
    break
  elif [ "$P4D_FLAG" = "true" ]; then
    echo "P4D_FLAG is true. Re-starting p4d server."
    eval "$RUN_COMMAND"
  else
  fi
  sleep 60
done

# START STOP REPLICATE
# DB UPGRADE
# DB INTEGRITY CHECK
# BACKUP (OPTIONAL
# https://help.perforce.com/helix-core/server-apps/p4sag/current/Content/P4SAG/appendix.p4d.html#P4_Server_(p4d)_reference)