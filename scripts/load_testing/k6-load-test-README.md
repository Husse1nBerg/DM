# K6 Load Testing Script for Marina Management System

This k6 script is designed to perform load testing on the marina management system backend, focusing on read operations while excluding customer, vessel, email, and SMS endpoints.

## Prerequisites

1. **Install k6**: Follow the installation guide at [k6.io](https://k6.io/docs/getting-started/installation/)
2. **Running Backend**: Ensure your marina management system backend is running
3. **Test User**: Have a test user account with appropriate permissions

## Configuration

### Environment Variables

Set the following environment variables before running the test:

```bash
export BASE_URL="http://localhost:8080"          # Your backend URL
export TEST_EMAIL="your-test-user@example.com"  # Test user email
export TEST_PASSWORD="your-test-password"       # Test user password
```

### Test Configuration

The script includes several configurable parameters in the `options` object:

- **Load Stages**: Currently configured for a 16-minute test with gradual ramp-up
- **Thresholds**: 
  - 95% of requests should complete in under 2 seconds
  - Error rate should be less than 10%

## Running the Tests

### Basic Test Run
```bash
k6 run k6-load-test.js
```

### With Environment Variables
```bash
BASE_URL="https://your-api-domain.com" TEST_EMAIL="test@example.com" TEST_PASSWORD="password123" k6 run k6-load-test.js
```

### Different Load Patterns

#### Quick Test (Low Load)
```bash
k6 run --stage 1m:5,2m:5,1m:0 k6-load-test.js
```

#### Stress Test (High Load)
```bash
k6 run --stage 5m:50,10m:50,5m:0 k6-load-test.js
```

#### Spike Test
```bash
k6 run --stage 2m:10,1m:100,2m:10,1m:0 k6-load-test.js
```

## Tested Endpoints

The script tests the following read endpoints:

### Core Endpoints
- `GET /api/v1/health` - Health check
- `GET /api/v1/project-details` - Project details

### User Management
- `GET /api/v1/user/profile` - Current user profile
- `GET /api/v1/user/list` - List users (paginated)
- `GET /api/v1/user/role/:roleId` - Users by role
- `GET /api/v1/user/organization/:organizationId` - Users by organization
- `GET /api/v1/user/marina/:marinaId` - Users by marina

### Organizations
- `GET /api/v1/organizations` - List organizations (paginated)
- `GET /api/v1/organizations/by-email` - Organization by email
- `GET /api/v1/organizations/:id` - Organization by ID
- `GET /api/v1/organizations/:id/with-address` - Organization with address

### Marinas
- `GET /api/v1/marinas` - List marinas (paginated)
- `GET /api/v1/marinas/user` - Current user's marinas
- `GET /api/v1/marinas/by-email` - Marina by email
- `GET /api/v1/marinas/:id` - Marina by ID
- `GET /api/v1/marinas/:id/with-address` - Marina with address
- `GET /api/v1/marinas/:id/contacts` - Marina contacts

### Roles
- `GET /api/v1/role/list` - List roles
- `GET /api/v1/role/:roleId` - Role by ID
- `GET /api/v1/role/name` - Role by name

### Addresses
- `GET /api/v1/addresses/:id` - Address by ID

### DME Integration
- `GET /api/v1/dme/sysids` - List DME System IDs
- `GET /api/v1/dme/sysids/:id` - DME System ID by ID
- `GET /api/v1/dme/credentials/organization/:organizationId` - DME credentials

### Gallery
- `GET /api/v1/gallery/marina/:marinaId` - Marina gallery
- `GET /api/v1/gallery/marina/item/:id` - Marina gallery item

### Documents
- `GET /api/v1/documents/user` - User documents
- `GET /api/v1/documents/:id` - Document by ID

### Work Orders
- `GET /api/v1/work-orders/list` - List work orders
- `GET /api/v1/work-orders/search` - Search work orders
- `GET /api/v1/work-orders/operations` - Work order operations

### Messages
- `GET /api/v1/message/marina` - Marina messages
- `GET /api/v1/message/get` - Message by ID

### Marina Usage History
- `GET /api/v1/marina-usage-history/marina/:marinaId` - Marina usage history
- `GET /api/v1/marina-usage-history/marina/:marinaId/latest` - Latest usage history

### Plans
- `GET /api/v1/plans/notes-messages` - Notes and messages plans
- `GET /api/v1/plans/storage` - Storage plans

### Permissions (Testing)
- `GET /api/v1/test/permissions/user` - User permissions
- `GET /api/v1/test/permissions/routes` - Route permissions

## Excluded Endpoints

As requested, the following endpoints are **NOT** included in the load test:

- Customer endpoints (`/customers/*`)
- Vessel/Boat endpoints (`/boats/*`)
- Email endpoints (`/email/*`)
- SMS endpoints (`/sms/*`)

## Metrics and Monitoring

The script tracks the following metrics:

- **HTTP Request Duration**: Response times for all requests
- **HTTP Request Failed**: Failed request rate
- **Custom Error Rate**: Additional error tracking
- **Response Status Codes**: Success/failure breakdown

## Interpreting Results

### Key Metrics to Monitor:
- **Average Response Time**: Should be under 1 second for most endpoints
- **95th Percentile Response Time**: Should be under 2 seconds
- **Error Rate**: Should be under 10%
- **Requests Per Second**: Shows throughput capability

### Common Issues:
- **High Error Rates**: Check authentication, permissions, or server capacity
- **Slow Response Times**: May indicate database performance issues
- **Authentication Failures**: Verify test credentials and token expiration

## Customization

### Adjusting Load Patterns
Modify the `stages` array in the `options` object:

```javascript
stages: [
  { duration: '5m', target: 20 },  // Ramp up to 20 users over 5 minutes
  { duration: '10m', target: 20 }, // Stay at 20 users for 10 minutes
  { duration: '3m', target: 0 },   // Ramp down to 0 users
],
```

### Adding New Endpoints
Add new test functions following the existing pattern:

```javascript
function testNewEndpoints(headers) {
  const response = http.get(`${BASE_API_URL}/new-endpoint`, { headers });
  const success = check(response, {
    'new endpoint status is 200': (r) => r.status === 200,
  });
  errorRate.add(!success);
}
```

### Modifying Thresholds
Adjust performance expectations in the `thresholds` object:

```javascript
thresholds: {
  http_req_duration: ['p(95)<3000'], // Allow 3 seconds for 95th percentile
  http_req_failed: ['rate<0.05'],    // Allow 5% error rate
},
```

## Troubleshooting

### Common Issues:

1. **Authentication Errors**: 
   - Verify `TEST_EMAIL` and `TEST_PASSWORD` are correct
   - Check if the test user has appropriate permissions
   - Ensure the user account is active

2. **Connection Errors**:
   - Verify `BASE_URL` is correct and accessible
   - Check if the backend service is running
   - Verify network connectivity

3. **Permission Errors**:
   - Ensure the test user has access to the tested endpoints
   - Check RBAC permissions in the system
   - Verify the user is assigned to appropriate marinas/organizations

4. **High Error Rates**:
   - Check server logs for specific error messages
   - Monitor database performance
   - Verify server capacity can handle the load

### Debug Mode
Run with verbose output to see detailed request/response information:

```bash
k6 run --http-debug k6-load-test.js
```

## Best Practices

1. **Start Small**: Begin with low load and gradually increase
2. **Monitor Resources**: Watch CPU, memory, and database performance on the server
3. **Test Realistic Scenarios**: Use actual data patterns and user behaviors
4. **Regular Testing**: Run load tests regularly, especially before deployments
5. **Environment Isolation**: Use dedicated test environments when possible 