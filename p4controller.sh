#!/bin/sh

ENV P4D-IP="0.0.0.0"

p4d -V

p4d -r /var/p4d-root/ -L log -p $P4D_IP:4232 -d

# Keep container running
while true; do sleep 60; done 