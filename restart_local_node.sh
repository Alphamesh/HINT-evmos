#!/bin/bash

ps aux | grep local_node | grep -v grep|cut -c 9-16|xargs kill -15
ps aux | grep 'evmosd start' | grep -v grep|cut -c 9-16|xargs kill -15
sleep 10
/home/ubuntu/code/HINT-evmos/local_node.sh