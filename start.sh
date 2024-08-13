#!/bin/bash

echo "Starting backend..."
cd backend
go run main.go &
BACKEND_PID=$!

cd ..
echo "Starting frontend..."
cd frontend
npm start &
FRONTEND_PID=$!

# Wait for both processes to finish
wait $BACKEND_PID
wait $FRONTEND_PID