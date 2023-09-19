#!/bin/bash

cd /home/ubuntu/code/HINT-evmos
ps aux | grep 'evmosd start' | grep -v grep|cut -c 9-16|xargs kill -15
sleep 10
./ubuntu_local_node.sh