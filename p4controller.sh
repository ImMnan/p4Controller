#!/bin/sh

pwd

ls -l 

p4d -V

p4d -r /var/p4d-root/ -L log -p 0.0.0.0:4232 -d

# Keep container running
while true; do sleep 60; done