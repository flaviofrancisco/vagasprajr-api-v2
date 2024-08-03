#!/bin/bash

# Load variables from .env file
if [ -f .env ]; then
  source .env
else
  echo ".env file not found!"
  exit 1
fi

# Log in to Docker Hub
echo "Logging in to Docker Hub..."
docker login -u $DOCKER_HUB_USERNAME

# Build the Docker image
echo "Building the Docker image..."
docker-compose build

# Tag the Docker image
echo "Tagging the Docker image..."
docker tag ${IMAGE_NAME}:${TAG} ${DOCKER_HUB_USERNAME}/${DOCKER_HUB_REPO}:${TAG}

# Push the Docker image to Docker Hub
echo "Pushing the Docker image to Docker Hub..."
docker push ${DOCKER_HUB_USERNAME}/${DOCKER_HUB_REPO}:${TAG}

echo "Docker image published successfully!"