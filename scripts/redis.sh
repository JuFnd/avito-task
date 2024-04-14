#!/bin/bash

# Install necessary dependencies
apt-get update
apt-get install curl gpg sudo deb -y

# Import Redis GPG key
curl -fsSL https://packages.redis.io/gpg | sudo gpg --dearmor -o /usr/share/keyrings/redis-archive-keyring.gpg

# Add Redis repository to sources.list.d
echo "deb [signed-by=/usr/share/keyrings/redis-archive-keyring.gpg] https://packages.redis.io/deb $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/redis.list

# Update package lists
apt-get update

# Install Redis
apt-get install redis -y
