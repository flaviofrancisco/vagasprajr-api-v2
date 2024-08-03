#!/bin/bash

# Navigate to the specified directory
cd /home/flavio/workspace/vagasprajr-api-v2

# Source the environment file
source /etc/environment

# Pull the latest changes from the main branch
git pull origin main

# Build the Docker images
docker-compose build

# Stop any running containers
docker-compose down

# Start the containers in detached mode
docker-compose up -d