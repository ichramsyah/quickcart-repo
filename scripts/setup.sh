#!/bin/bash

# QuickCart - One-Command Setup Script
# This script automates the entire local deployment process.

set -e

echo "🚀 Starting QuickCart Setup..."

# Check Prerequisites
if ! command -v docker &> /dev/null; then
    echo "❌ Docker could not be found. Please install Docker first."
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose could not be found. Please install Docker Compose first."
    exit 1
fi

# Navigate to Part 2 Project Folder (Infrastructure)
echo "📂 Navigating to Infrastructure folder..."
cd ../part2-infrastructure/project_folder

# Create .env if not exists
if [ ! -f .env ]; then
    echo "📝 Creating .env from .env.example..."
    cp .env.example .env
fi

# Build & Start Containers
echo "whale: Building and Starting Services..."
docker-compose up -d --build

# Wait for Healthchecks
echo "zzz Waiting for Database to be ready..."
until docker-compose exec postgres pg_isready -U quickcart; do
  echo "   Waiting for postgres..."
  sleep 2
done

echo "✅ Database is ready!"

# Final Status
echo ""
echo "🎉 Setup Complete! QuickCart is running."
echo "----------------------------------------"
echo "🌍 App URL:      http://localhost:8080"
echo "📊 Database:     localhost:5432"
echo "🧠 Redis:        localhost:6379"
echo "----------------------------------------"
echo "To stop the app, run: ./scripts/teardown.sh"
