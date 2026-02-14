#!/bin/bash

# QuickCart - cleanup Script
# This script removes containers, networks, and images created by setup.sh.

echo "Starting Cleanup..."

# Navigate to Part 2 Project Folder
cd ../part2-infrastructure/project_folder

# Stop and remove containers, networks
echo "Stopping containers..."
docker-compose down

# Optional: Remove volumes (Ask user)
read -p "Do you want to remove persistent volumes (Database data will be lost)? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🗑️ Removing volumes..."
    docker-compose down -v
fi

echo "✅ Cleanup Complete!"
