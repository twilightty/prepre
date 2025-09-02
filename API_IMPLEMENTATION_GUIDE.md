# ATMT API Implementation Guide

## Overview

ATMT (Automation Tool Management) is a complete payment and product distribution system with:
- JWT-based authentication
- SePay webhook payment processing  
- Secure file downloads with serial validation
- 8-character payment code generation

**Base URL**: `http://localhost:8080`
**API Version**: v1

## Quick Start Implementation

### 1. Authentication Flow

#### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "full_name": "John Doe",
  "date_of_birth": "1990-01-01T00:00:00Z",
  "platform": "windows",
  "serial_number": "USER001"
}
```

**Response:**
```json
{
  "user": {
    "id": "66f123abc456def789012345",
    "email": "user@example.com",
    "full_name": "John Doe",
    "date_of_birth": "1990-01-01T00:00:00Z",
    "platform": "windows",
    "owned": false,
    "is_banned": false,
    "serial_number": "USER001",
    "role": "user",
    "is_active": true,
    "created_at": "2025-09-02T15:30:00Z",
    "updated_at": "2025-09-02T15:30:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": 1725292200
}
```

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:** Same structure as register

### 2. Payment Flow Implementation

#### Step 1: Initiate Payment
```http
POST /api/v1/payment/initiate
Authorization: Bearer YOUR_JWT_TOKEN
```

**Response:**
```json
{
  "payment_code": "ABC12345",
  "amount": 5000000,
  "qr_image_url": "https://img.vietqr.io/image/mbbank-28368866886-compact.jpg?amount=5000000&addInfo=ATMTABC12345&accountName=NGUYEN%20HONG%20QUANG",
  "expires_at": "2025-09-02T15:45:00Z",
  "message": "Please scan the QR code to complete payment. Payment code: ABC12345"
}
```

#### Step 2: Display QR Code to User
Show the `qr_image_url` to the user for bank transfer. Payment details:
- **Bank**: MB Bank (Military Bank)
- **Account**: 28368866886 
- **Account Name**: NGUYEN HONG QUANG
- **Amount**: 5,000,000 VND (exactly)
- **Transfer Note**: ATMT{8-character-code}

#### Step 3: SePay Webhook (Automatic)
When user completes payment, SePay sends webhook:

```http
POST /hooks/sepay
Authorization: ApiKey xoxoxoxoxoxo
Content-Type: application/json

{
  "id": 12345,
  "gateway": "MB Bank",
  "transactionDate": "2025-09-02 15:35:42",
  "accountNumber": "28368866886",
  "content": "ATMTABC12345 chuyen tien mua san pham",
  "transferType": "in",
  "transferAmount": 5000000,
  "accumulated": 15000000,
  "referenceCode": "REF123456",
  "description": "Incoming transfer"
}
```

**System automatically:**
1. Validates amount = 5,000,000 VND
2. Extracts payment code from content (ATMT + 8 chars)
3. Finds user with matching payment code
4. Sets user `owned: true`
5. Enables product downloads

### 3. Product Downloads

#### List Available Products
```http
GET /api/v1/products
Authorization: Bearer YOUR_JWT_TOKEN
```

**Response:**
```json
{
  "products": [
    {
      "name": "chatgpt",
      "display_name": "ChatGPT",
      "available": true,
      "platforms": ["windows", "macos"]
    },
    {
      "name": "dalle",
      "display_name": "DALL-E", 
      "available": true,
      "platforms": ["windows", "macos"]
    }
  ],
  "user": {
    "email": "user@example.com",
    "full_name": "John Doe",
    "platform": "windows",
    "owned": true,
    "serial_number": "USER001"
  }
}
```

#### Download Product
```http
GET /api/v1/download/{product_name}/{platform}?serial={serial}
Authorization: Bearer YOUR_JWT_TOKEN
```

**Examples:**
```bash
# Download ChatGPT for Windows
GET /api/v1/download/chatgpt/windows?serial=USER001

# Download DALL-E for macOS  
GET /api/v1/download/dalle/macos?serial=USER001
```

**Products Available:**
- `chatgpt` - ChatGPT
- `dalle` - DALL-E
- `gemini` - Google Gemini
- `hailuo` - Hailuo AI
- `runway` - Runway ML
- `sora` - OpenAI Sora
- `veo3` - Google Veo 3
- `veo3_pro` - Google Veo 3 Pro

**Platforms:** `windows`, `macos`

## Implementation Examples

### Frontend Integration

#### JavaScript/React Example
```javascript
class ATMTClient {
  constructor(baseUrl = 'http://localhost:8080') {
    this.baseUrl = baseUrl;
    this.token = localStorage.getItem('atmt_token');
  }

  async register(userData) {
    const response = await fetch(`${this.baseUrl}/api/v1/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(userData)
    });
    
    const data = await response.json();
    if (data.token) {
      this.token = data.token;
      localStorage.setItem('atmt_token', data.token);
    }
    return data;
  }

  async initiatePayment() {
    const response = await fetch(`${this.baseUrl}/api/v1/payment/initiate`, {
      method: 'POST',
      headers: { 
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json'
      }
    });
    return response.json();
  }

  async downloadProduct(productName, platform, serial) {
    const response = await fetch(
      `${this.baseUrl}/api/v1/download/${productName}/${platform}?serial=${serial}`, 
      {
        headers: { 'Authorization': `Bearer ${this.token}` }
      }
    );
    
    if (response.ok) {
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${productName}.exe`;
      a.click();
    }
  }
}
```

#### Usage Example
```javascript
const client = new ATMTClient();

// Register user
await client.register({
  email: 'user@example.com',
  password: 'password123',
  full_name: 'John Doe',
  date_of_birth: '1990-01-01T00:00:00Z',
  platform: 'windows',
  serial_number: 'USER001'
});

// Initiate payment
const payment = await client.initiatePayment();
console.log('Payment QR:', payment.qr_image_url);
console.log('Payment Code:', payment.payment_code);

// After payment completion, download products
await client.downloadProduct('chatgpt', 'windows', 'USER001');
```

### Backend Webhook Handler

#### Node.js Express Example
```javascript
app.post('/webhook/sepay', (req, res) => {
  const apiKey = req.headers.authorization;
  
  if (apiKey !== 'ApiKey xoxoxoxoxoxo') {
    return res.status(401).json({ error: 'Invalid API key' });
  }

  const webhookData = req.body;
  
  // Forward to ATMT system
  fetch('http://localhost:8080/hooks/sepay', {
    method: 'POST',
    headers: {
      'Authorization': apiKey,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(webhookData)
  });

  res.json({ success: true });
});
```

#### PHP Example
```php
<?php
// Webhook endpoint
if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $headers = getallheaders();
    $auth = $headers['Authorization'] ?? '';
    
    if ($auth !== 'ApiKey xoxoxoxoxoxo') {
        http_response_code(401);
        exit(json_encode(['error' => 'Invalid API key']));
    }
    
    $webhookData = json_decode(file_get_contents('php://input'), true);
    
    // Forward to ATMT
    $ch = curl_init('http://localhost:8080/hooks/sepay');
    curl_setopt($ch, CURLOPT_POST, 1);
    curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($webhookData));
    curl_setopt($ch, CURLOPT_HTTPHEADER, [
        'Authorization: ApiKey xoxoxoxoxoxo',
        'Content-Type: application/json'
    ]);
    curl_exec($ch);
    curl_close($ch);
    
    echo json_encode(['success' => true]);
}
?>
```

## Error Handling

### Common HTTP Status Codes
```
200 - Success
400 - Bad Request (validation error)
401 - Unauthorized (invalid/missing token)
403 - Forbidden (insufficient permissions)
404 - Not Found
409 - Conflict (user already exists)
500 - Internal Server Error
```

### Error Response Format
```json
{
  "error": "Error message description",
  "code": 400
}
```

### Specific Error Cases

#### Authentication Errors
- `"Authorization header required"` - Missing Bearer token
- `"Invalid or expired token"` - Token validation failed
- `"Invalid email or password"` - Login credentials incorrect

#### Payment Errors
- `"User already owns the product"` - User has already purchased
- `"User is banned and cannot make payments"` - Account suspended
- `"Incorrect payment amount: expected 5000000, got X"` - Wrong amount transferred
- `"Payment code not found in content"` - Transfer note missing ATMT code

#### Download Errors
- `"You do not own this product"` - Payment not completed
- `"Serial number does not match your account"` - Serial validation failed
- `"Invalid product name"` - Product doesn't exist
- `"Product file not found"` - File missing on server

## Security Implementation

### JWT Token Structure
```javascript
// Token payload
{
  "user_id": "66f123abc456def789012345",
  "email": "user@example.com", 
  "role": "user",
  "type": "access",
  "iat": 1725288000,
  "exp": 1725374400
}
```

### API Key Validation
SePay webhook endpoint requires exact header:
```
Authorization: ApiKey xoxoxoxoxoxo
```

### Serial Number Validation
- Must match user's registered serial exactly
- Required for all download requests
- Cannot be changed after registration

## Testing & Development

### Test Scripts Available
```bash
# Test authentication flow
./scripts/test_auth.sh

# Test payment system
./scripts/test_payment.sh  

# Test download system
./scripts/test_download.sh
```

### Manual Testing Commands
```bash
# Register user
curl -X POST "http://localhost:8080/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123", 
    "full_name": "Test User",
    "platform": "windows",
    "serial_number": "TEST001"
  }'

# Initiate payment
curl -X POST "http://localhost:8080/api/v1/payment/initiate" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Simulate webhook
curl -X POST "http://localhost:8080/hooks/sepay" \
  -H "Authorization: ApiKey xoxoxoxoxoxo" \
  -H "Content-Type: application/json" \
  -d '{
    "id": 12345,
    "transferAmount": 5000000,
    "content": "ATMTTEST1234 payment",
    "transferType": "in"
  }'

# Download product
curl -X GET "http://localhost:8080/api/v1/download/chatgpt/windows?serial=TEST001" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  --output chatgpt.exe
```

## Deployment Checklist

### Environment Configuration
- [ ] Update `config.yaml` with production settings
- [ ] Set MongoDB connection string
- [ ] Configure SePay API key
- [ ] Set JWT secret key
- [ ] Configure CORS origins

### File Structure Setup
```
dist/
├── chatgpt/
│   ├── windows/chatgpt.exe
│   └── macos/chatgpt
├── dalle/
│   ├── windows/dalle.exe
│   └── macos/dalle
└── ...
```

### Security Considerations
- [ ] Use HTTPS in production
- [ ] Rotate JWT secrets regularly
- [ ] Monitor failed payment attempts
- [ ] Log all download activities
- [ ] Rate limit API endpoints

### Monitoring Setup
- [ ] Monitor webhook delivery status
- [ ] Track payment success rates
- [ ] Monitor download bandwidth
- [ ] Alert on failed authentications
- [ ] Log payment processing errors

This implementation guide provides everything needed to integrate ATMT into your system! 🚀
