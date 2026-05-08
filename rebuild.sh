#!/bin/bash

# SentiGraph Finance - Rebuild Script
# This script stops the current stack and rebuilds all Docker images from scratch.

echo "🛑 Stopping SentiGraph stack..."
docker-compose down

echo "🧹 Cleaning up dangling images..."
docker image prune -f

echo "🏗️  Rebuilding and starting services..."
# Using --build to force recreation of images
# Using --force-recreate to ensure fresh containers
docker-compose up -d --build --force-recreate

echo "✅ Stack is up and running!"
echo "📈 Dashboard: http://localhost:3000"
echo "🧠 Analyst API: http://localhost:8080"
echo ""
echo "💡 Use 'docker-compose logs -f' to monitor the services."
