#!/bin/bash
set -e

echo "==> Building frontend..."
npm install
npm run build

echo "==> Copying frontend to web/"
rm -rf web
cp -r build web
rm -rf build

echo "==> Building Go backend..."
cd backend
go build -o ../pos ./cmd

echo "==> Done! Run ./pos to start the server"
