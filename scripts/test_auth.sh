#!/bin/bash

# Authentication API Test Script
# Make sure the server is running before running this script

BASE_URL="http://localhost:8080/api/v1"

echo "=== Testing Authentication API ==="
echo

# Test 1: Register a new user
echo "1. Testing user registration..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User"
  }')

echo "Register Response: $REGISTER_RESPONSE"
echo

# Test 2: Login with the created user
echo "2. Testing user login..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }')

echo "Login Response: $LOGIN_RESPONSE"
echo

# Extract token from login response
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
REFRESH_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "Failed to extract token from login response"
  exit 1
fi

echo "Extracted Token: $TOKEN"
echo "Extracted Refresh Token: $REFRESH_TOKEN"
echo

# Test 3: Get user profile with token
echo "3. Testing authenticated profile access..."
PROFILE_RESPONSE=$(curl -s -X GET "$BASE_URL/auth/profile" \
  -H "Authorization: Bearer $TOKEN")

echo "Profile Response: $PROFILE_RESPONSE"
echo

# Test 4: Test token refresh
echo "4. Testing token refresh..."
REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\": \"$REFRESH_TOKEN\"}")

echo "Refresh Response: $REFRESH_RESPONSE"
echo

# Test 5: Test logout
echo "5. Testing logout..."
LOGOUT_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/logout" \
  -H "Authorization: Bearer $TOKEN")

echo "Logout Response: $LOGOUT_RESPONSE"
echo

# Test 6: Test invalid token access
echo "6. Testing access with invalid token..."
INVALID_RESPONSE=$(curl -s -X GET "$BASE_URL/auth/profile" \
  -H "Authorization: Bearer invalid_token")

echo "Invalid Token Response: $INVALID_RESPONSE"
echo

# Test 7: Test health endpoint
echo "7. Testing health endpoint..."
HEALTH_RESPONSE=$(curl -s -X GET "http://localhost:8080/health")

echo "Health Response: $HEALTH_RESPONSE"
echo

echo "=== Authentication API Tests Complete ==="
