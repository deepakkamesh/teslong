#!/bin/sh
GOOS=linux GOARCH=arm GOARM=6 go build
rsync -avz -e "ssh -o StrictHostKeyChecking=no" --progress teslong 192.168.0.108:~/

