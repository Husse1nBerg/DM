# SADIE Integration Guide - Complete Step-by-Step Tutorial

## Table of Contents
1. [Introduction](#introduction)
2. [What is SADIE?](#what-is-sadie)
3. [Project Setup](#project-setup)
4. [Adding the Phone Numbers Endpoint](#adding-the-phone-numbers-endpoint)
5. [Updating the Agent's Webhook URL](#updating-the-agents-webhook-url)
6. [How It All Works](#how-it-all-works)
7. [Testing the Endpoint](#testing-the-endpoint)
8. [Understanding the Code](#understanding-the-code)

---

## Introduction

This guide explains everything we did to add a SADIE phone numbers endpoint to the marina management backend. We'll walk through each step in simple terms, assuming you have no prior coding experience.

---

## What is SADIE?

**SADIE** is an AI-powered voice assistant system that can handle phone calls. In this project, SADIE is integrated to:
- Answer phone calls from customers
- Check for available boat slips (parking spaces for boats)
- Make reservations
- Interact with customers through natural conversation

Think of SADIE as a smart receptionist that never sleeps and can handle multiple calls at once.

---

## Project Setup

Before we could add new features, we needed to set up the project properly.

### Step 1: Creating the Environment File (.env)

**What is a .env file?**
A `.env` file is like a secret notebook that stores important configuration information. It contains things like:
- Database passwords
- API keys
- Server settings

**What we did:**
We created a `.env` file with the following information:

```
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=marina_dev

# Server Configuration
PORT=8080
HOST=localhost
ENV=development

# Auth Configuration
ACCESS_SECRET=your_access_secret_here
REFRESH_SECRET=your_refresh_secret_here
```

**Why this matters:**
The application needs to know where to find the database and how to connect to it. Without this file, the server wouldn't know where to look for data.

### Step 2: Starting the Database

**What is a database?**
A database is like a digital filing cabinet that stores all the information your application needs (users, boats, reservations, etc.).

**What we did:**
We started a PostgreSQL database using Docker (a tool that runs applications in isolated containers):

```bash
docker compose -f docker-compose.dev.yml up -d
```

This command:
- Starts a database server
- Makes it available on port 5432
- Creates a database called `marina_dev`

### Step 3: Running Database Migrations

**What are migrations?**
Migrations are like blueprints that create the structure of your database. They define what tables exist and what information each table can store.

**What we did:**
We ran migrations to set up all the database tables:

```bash
# We used a workaround because 'make' wasn't available on Windows
docker cp db/migrations dm-web-backend-db-1:/tmp/migrations
docker exec dm-web-backend-db-1 sh -c "apk add --no-cache wget && cd /tmp && wget -q https://github.com/pressly/goose/releases/download/v3.21.1/goose_linux_x86_64 -O goose && chmod +x goose && ./goose -dir /tmp/migrations postgres 'postgres://postgres:postgres@localhost:5432/marina_dev?sslmode=disable' up"
```

**What happened:**
- 44 migration files were executed
- All database tables were created
- The database is now ready to store data

### Step 4: Generating Database Code

**What is SQLC?**
SQLC is a tool that automatically creates Go code from SQL queries. Instead of writing database code manually, SQLC generates it for us.

**What we did:**
```bash
sqlc generate
```

This created Go functions that we can use to interact with the database.

------------------------------------------------------------------------------------------------------------------------------------------------

## Adding the Phone Numbers Endpoint

Now let's get to the main part: adding the endpoint to retrieve phone numbers from SADIE.

### What is an Endpoint?

An **endpoint** is like a specific address in your application. Just like:
- `https://google.com` takes you to Google's homepage
- `https://google.com/search` takes you to Google's search page

An endpoint in our API is a specific URL that does a specific thing:
- `http://localhost:8080/getPhoneNumbers` - Gets all phone numbers
- `http://localhost:8080/getAssistantPhoneNumber` - Gets a specific assistant's phone number
- `http://localhost:8080/updateAgentWebhook` - Updates where SADIE sends webhook events
- `http://localhost:8080/webhooks` - Receives webhook events from SADIE

### Step 1: Adding the Function to Get Phone Numbers from SADIE API

**File:** `pkg/sadie/client.go`

**What we added:**
We created a new function called `GetPhoneNumbers()` that talks to the SADIE API and asks for all available phone numbers.

```go
// GetPhoneNumbers retrieves all phone numbers for the tenant
func (c *Client) GetPhoneNumbers(ctx context.Context) ([]PhoneNumber, error) {
    // This function makes a GET request to /phone-numbers endpoint
    // and returns a list of phone numbers
}
```

**In simple terms:**
- This function is like a phone call to SADIE's API
- It says: "Hey SADIE, give me all the phone numbers you have"
- SADIE responds with a list of phone numbers
- The function returns that list to us

### Step 2: Creating the Handler Function

**File:** `internal/server/handlers/sadie_handlers.go`

**What is a handler?**
A handler is like a waiter in a restaurant:
- You (the customer) make a request (order food)
- The waiter (handler) takes your request
- The waiter goes to the kitchen (does some work)
- The waiter brings back your food (returns a response)

**What we added:**
We created a handler function called `GetPhoneNumbersHandler()`:

```go
func (h *SadieHandler) GetPhoneNumbersHandler(c echo.Context) error {
    // 1. Create a SADIE client (like picking up the phone)
    sadieClient := sadie.NewClient(&h.server.Config.Sadie, h.server.Logger)
    
    // 2. Call the GetPhoneNumbers function (make the phone call)
    phoneNumbers, err := sadieClient.GetPhoneNumbers(c.Request().Context())
    
    // 3. If there's an error, return an error message
    if err != nil {
        return responses.NewSadieErrorResponse(...)
    }
    
    // 4. Format the response and send it back
    response := responses.NewSadieSuccessResponse(...)
    return response.JSON(c)
}
```

**Step-by-step what happens:**
1. Someone makes a request to `/getPhoneNumbers`
2. The handler receives the request
3. The handler creates a SADIE client (like opening a communication channel)
4. The handler calls SADIE's API to get phone numbers
5. The handler formats the response nicely
6. The handler sends the response back to the person who asked

### Step 3: Registering the Route

**File:** `internal/server/routes/sadie_routes.go`

**What is a route?**
A route is like a map that tells the server: "When someone visits this URL, run this function."

**What we added:**
We added this line to register our new endpoint:

```go
sadieProtected.GET("/getPhoneNumbers", sadieHandler.GetPhoneNumbersHandler)
```

**What this means:**
- `sadieProtected` - This is a group of routes that require authentication
- `.GET` - This endpoint only responds to GET requests (like typing a URL in your browser)
- `"/getPhoneNumbers"` - This is the URL path
- `sadieHandler.GetPhoneNumbersHandler` - This is the function to run when someone visits that URL

**The full picture:**
```
User makes request → Server checks routes → Finds matching route → Runs handler function → Returns response
```

---

## Updating the Agent's Webhook URL

Now that we have endpoints for SADIE to call, we need to tell SADIE where to find them. This is what we call "updating the webhook URL."

### What is a Webhook URL?

Think of a **webhook URL** like a mailing address:
- When something important happens (like a phone call), SADIE needs to know where to send that information
- The webhook URL is like the address where SADIE should "mail" (send) information about phone calls
- Our server has a special endpoint at `/webhooks` that receives this information

**Why do we need to update it?**
- When you first set up a SADIE assistant, it might not know where your server is
- Or your server address might change (like moving from development to production)
- We need to tell SADIE: "Hey, when something happens, send the information to this address: `http://localhost:8080/webhooks`"

### Step 1: Adding Webhook URL Configuration

**File:** `internal/config/sadie.go`

**What we added:**
We added a new field called `WebhookURL` to store the address where SADIE should send webhook events.

**In simple terms:**
- We're telling our application: "Remember this address - this is where SADIE should send information"
- The application automatically figures out the address based on your server settings
- For development, it might be: `http://localhost:8080/webhooks`
- For production, it might be: `https://your-domain.com/webhooks`

**What the code does:**
```go
// If webhook URL is not set, try to construct it from server config
if webhookURL == "" {
    // Automatically builds the URL like: http://localhost:8080/webhooks
    webhookURL = fmt.Sprintf("http://%s:%s/webhooks", serverHost, serverPort)
}
```

**Why this is helpful:**
- You don't have to manually set the webhook URL every time
- It automatically uses the right address (http for development, https for production)
- You can still override it by setting `SADIE_WEBHOOK_URL` in your `.env` file if needed

### Step 2: Adding the Function to Update Webhook URL in SADIE

**File:** `pkg/sadie/client.go`

**What we added:**
We created a new function called `UpdateAssistantWebhookURL()` that tells SADIE's API to update the webhook URL for an assistant.

```go
// UpdateAssistantWebhookURL updates the webhook URL for a specific assistant
func (c *Client) UpdateAssistantWebhookURL(ctx context.Context, assistantID, webhookURL string) error {
    // This function makes a PATCH request to SADIE's API
    // It says: "Hey SADIE, update this assistant's webhook URL to this new address"
}
```

**In simple terms:**
- This function is like calling SADIE's customer service
- It says: "Hi, I need to update the mailing address for assistant #123"
- It gives SADIE the new address: `http://localhost:8080/webhooks`
- SADIE updates its records and says "Got it!"

**What PATCH means:**
- **PATCH** is like an update request - "Change this one thing"
- It's different from POST (create new) or GET (read information)
- We're updating just the webhook URL, not changing anything else about the assistant

### Step 3: Creating the Handler Function

**File:** `internal/server/handlers/sadie_handlers.go`

**What we added:**
We created a handler function called `UpdateAgentWebhookHandler()` that handles requests to update the webhook URL.

**What is a handler?**
Remember, a handler is like a waiter in a restaurant - it takes your request, does the work, and brings back a response.

**What this handler does:**
```go
func (h *SadieHandler) UpdateAgentWebhookHandler(c echo.Context) error {
    // 1. Figure out which assistant to update (or use the first one)
    // 2. Get the webhook URL (use configured one or from request)
    // 3. Call SADIE's API to update it
    // 4. Return success or error message
}
```

**Step-by-step what happens:**
1. Someone makes a request to `/updateAgentWebhook`
2. The handler receives the request
3. The handler figures out which assistant to update:
   - If an `assistant_id` is provided, use that one
   - If not, get all assistants and use the first one
4. The handler gets the webhook URL:
   - If provided in the request, use that
   - Otherwise, use the configured webhook URL from settings
5. The handler calls SADIE's API to update the webhook URL
6. The handler sends back a success message or an error

**Why this is smart:**
- You don't have to remember the assistant ID - it can find it automatically
- You don't have to specify the webhook URL - it uses the configured one
- But you CAN override both if you want to

### Step 4: Registering the Route

**File:** `internal/server/routes/sadie_routes.go`

**What we added:**
We added this line to register our new endpoint:

```go
sadieProtected.POST("/updateAgentWebhook", sadieHandler.UpdateAgentWebhookHandler)
```

**What this means:**
- `sadieProtected` - This route requires authentication (the `x-sadie-core-secret` header)
- `.POST` - This endpoint responds to POST requests (sending data)
- `"/updateAgentWebhook"` - This is the URL path
- `sadieHandler.UpdateAgentWebhookHandler` - This is the function to run

**The full picture:**
```
You make POST request → Server checks authentication → Finds route → Runs handler → Handler updates SADIE → Returns success
```

### How the Webhook URL is Determined

The system is smart about figuring out the webhook URL:

1. **First, it checks if you set it manually:**
   - Look for `SADIE_WEBHOOK_URL` in your `.env` file
   - If found, use that exact URL

2. **If not set, it builds it automatically:**
   - For development: `http://localhost:8080/webhooks`
   - For production: `https://your-domain.com/webhooks`
   - It uses your server's host and port settings

3. **You can also provide it in the request:**
   - Send it in the request body: `{"webhook_url": "https://custom-url.com/webhooks"}`
   - This lets you override the configured URL for one-time updates

### Testing the Update Endpoint

**Using curl (command line):**

**Option 1: Simple update (uses first assistant and configured webhook URL):**
```bash
curl -X POST "http://localhost:8080/updateAgentWebhook" \
  -H "x-sadie-core-secret: dev_dockmaster_0195c925-ff69-7298-81a7-80f9348b52e6" \
  -H "Content-Type: application/json"
```

**Option 2: Update specific assistant:**
```bash
curl -X POST "http://localhost:8080/updateAgentWebhook?assistant_id=your-assistant-id" \
  -H "x-sadie-core-secret: dev_dockmaster_0195c925-ff69-7298-81a7-80f9348b52e6"
```

**Option 3: Update with custom webhook URL:**
```bash
curl -X POST "http://localhost:8080/updateAgentWebhook" \
  -H "x-sadie-core-secret: dev_dockmaster_0195c925-ff69-7298-81a7-80f9348b52e6" \
  -H "Content-Type: application/json" \
  -d '{"assistant_id": "your-assistant-id", "webhook_url": "https://your-domain.com/webhooks"}'
```

**What you'll see if successful:**
```json
{
  "success": true,g
  "data": {
    "assistant_id": "0195c920-ba64-7378-a10c-6a7d63496c1f",
    "webhook_url": "http://localhost:8080/webhooks",
    "message": "Webhook URL updated successfully"
  },
  "description": "Agent webhook URL updated successfully",
  "steps": [
    "The agent's webhook URL has been updated in SADIE",
    "SADIE will now send webhook events to the configured URL",
    "Test by making a call to the assistant's phone number"
  ]
}
```

**What you'll see if there's an error:**
```json
{
  "success": false,
  "error": "Failed to update assistant webhook URL",
  "description": "Could not update the webhook URL in SADIE API",
  "steps": [
    "Verify the assistant_id is correct",
    "Check that the webhook URL is valid",
    "Ensure SADIE API credentials are valid"
  ],
  "data": {
    "error": "SADIE API request failed with status 404: ...",
    "assistant_id": "invalid-id",
    "webhook_url": "http://localhost:8080/webhooks"
  }
}
```

### Why This Matters

**Before updating the webhook URL:**
- SADIE doesn't know where to send information about phone calls
- Your server won't receive webhook events
- You can't see what's happening during calls

**After updating the webhook URL:**
- SADIE knows exactly where to send information
- Your server receives webhook events at `/webhooks`
- You can monitor calls, see what customers are asking, and respond accordingly

**Real-world example:**
Imagine you're running a restaurant:
- **Before:** Customers call, but you never know about it because the phone company doesn't have your address to send notifications
- **After:** You've given the phone company your address, so they send you a notification every time someone calls

---

## How It All Works

Let's trace through what happens when you make a request:

### The Complete Flow

1. **You make a request:**
   ```
   GET http://localhost:8080/getPhoneNumbers
   Header: x-sadie-core-secret: dev_dockmaster_0195c925-ff69-7298-81a7-80f9348b52e6
   ```

2. **Server receives the request:**
   - The server is listening on port 8080
   - It sees someone wants `/getPhoneNumbers`
   - It checks if authentication is needed

3. **Authentication check:**
   - The route is protected, so it needs the `x-sadie-core-secret` header
   - The middleware checks if the secret matches
   - If it matches, the request continues
   - If it doesn't match, an error is returned

4. **Route matching:**
   - The server looks at all registered routes
   - It finds: `GET /getPhoneNumbers` → `GetPhoneNumbersHandler`
   - It knows to run the `GetPhoneNumbersHandler` function

5. **Handler executes:**
   - The handler creates a SADIE client
   - The handler calls `GetPhoneNumbers()` on the client
   - The client makes an HTTP request to SADIE's API: `GET https://core-api.heysadie.ai/phone-numbers`
   - SADIE API responds with phone numbers

6. **Response formatting:**
   - The handler receives the phone numbers
   - It formats them into a nice JSON response
   - It includes helpful steps for the user

7. **Response sent back:**
   - The server sends the JSON response back to you
   - You see the phone numbers!

### Visual Flow Diagram

```
┌─────────┐
│  You    │
│ (Client)│
└────┬────┘
     │ 1. Makes GET request
     │    with auth header
     ▼
┌─────────────────┐
│  Your Server    │
│  (localhost)    │
└────┬────────────┘
     │ 2. Checks authentication
     │ 3. Finds route
     │ 4. Runs handler
     ▼
┌─────────────────┐
│ SADIE Client    │
│ (pkg/sadie)     │
└────┬────────────┘
     │ 5. Makes API call
     ▼
┌─────────────────┐
│  SADIE API      │
│ (heysadie.ai)   │
└────┬────────────┘
     │ 6. Returns phone numbers
     ▼
┌─────────────────┐
│  Handler        │
│  Formats data   │
└────┬────────────┘
     │ 7. Returns JSON response
     ▼
┌─────────┐
│  You    │
│ Receives│
│ Response│
└─────────┘
```

---

## Testing the Endpoint

### How to Make the Request

**Using curl (command line):**

```bash
curl -X GET "http://localhost:8080/getPhoneNumbers" \
  -H "x-sadie-core-secret: dev_dockmaster_0195c925-ff69-7298-81a7-80f9348b52e6"
```

**Breaking this down:**
- `curl` - A tool for making HTTP requests from the command line
- `-X GET` - This is a GET request (like visiting a webpage)
- `"http://localhost:8080/getPhoneNumbers"` - The URL we want to visit
- `-H` - This adds a header (extra information)
- `"x-sadie-core-secret: ..."` - The authentication secret

**What you'll see:**
```json
{
  "success": true,
  "data": {
    "count": 2,
    "phone_numbers": [
      {
        "id": "0196cb1f-9956-77da-9da1-431123f93975",
        "phoneNumber": "+18058371995"
      },
      {
        "id": "0196ac49-e5ef-71b5-9d70-82b646e29543",
        "phoneNumber": "+17272058008"
      }
    ]
  },
  "description": "Phone numbers retrieved successfully",
  "steps": [
    "Use one of these phone numbers to assign to an assistant",
    "Call the phone number to test your SADIE agent",
    "Monitor the webhook endpoints to see call events"
  ]
}
```

### Using Other Tools

**Postman:**
1. Open Postman
2. Create a new GET request
3. URL: `http://localhost:8080/getPhoneNumbers`
4. In Headers tab, add:
   - Key: `x-sadie-core-secret`
   - Value: `dev_dockmaster_0195c925-ff69-7298-81a7-80f9348b52e6`
5. Click Send

**Browser:**
- You can't easily add custom headers in a browser, so use curl or Postman instead

---

## Understanding the Code

Let's look at the actual code files and understand what each part does.

### File Structure

```
dm-web-backend/
├── pkg/
│   └── sadie/
│       └── client.go          ← Talks to SADIE API
├── internal/
│   ├── server/
│   │   ├── handlers/
│   │   │   └── sadie_handlers.go  ← Handles HTTP requests
│   │   └── routes/
│   │       └── sadie_routes.go    ← Maps URLs to handlers
│   └── config/
│       └── sadie.go               ← SADIE configuration
```

### 1. SADIE Client (`pkg/sadie/client.go`)

**Purpose:** This file contains code that communicates with the SADIE API.

**Key Function:**
```go
func (c *Client) GetPhoneNumbers(ctx context.Context) ([]PhoneNumber, error) {
    // Makes HTTP GET request to /phone-numbers
    // Returns list of phone numbers
}
```

**What it does:**
- Creates an HTTP request to SADIE's API
- Adds authentication headers
- Sends the request
- Receives the response
- Parses the JSON response into Go data structures
- Returns the phone numbers

### 2. Handler (`internal/server/handlers/sadie_handlers.go`)

**Purpose:** This file contains functions that handle incoming HTTP requests.

**Key Function:**
```go
func (h *SadieHandler) GetPhoneNumbersHandler(c echo.Context) error {
    // 1. Create SADIE client
    sadieClient := sadie.NewClient(&h.server.Config.Sadie, h.server.Logger)
    
    // 2. Get phone numbers from SADIE
    phoneNumbers, err := sadieClient.GetPhoneNumbers(c.Request().Context())
    
    // 3. Handle errors
    if err != nil {
        return responses.NewSadieErrorResponse(...)
    }
    
    // 4. Format and return response
    response := responses.NewSadieSuccessResponse(...)
    return response.JSON(c)
}
```

**What it does:**
- Receives the HTTP request
- Creates a SADIE client
- Calls the SADIE API
- Handles any errors
- Formats a nice response
- Sends the response back

### 3. Routes (`internal/server/routes/sadie_routes.go`)

**Purpose:** This file maps URLs to handler functions.

**Key Code:**
```go
sadieProtected.GET("/getPhoneNumbers", sadieHandler.GetPhoneNumbersHandler)
```

**What it does:**
- Tells the server: "When someone visits `/getPhoneNumbers`, run `GetPhoneNumbersHandler`"
- Makes sure the route is protected (requires authentication)

### 4. Configuration (`internal/config/sadie.go`)

**Purpose:** This file stores SADIE API configuration.

**What it contains:**
- Base URL: `https://core-api.heysadie.ai`
- API Key: Your SADIE API key
- Tenant ID: Your tenant identifier
- Client Secret: Your authentication secret
- Webhook URL: The address where SADIE should send webhook events (automatically configured)

### 5. Update Webhook Function (`pkg/sadie/client.go`)

**Purpose:** This function communicates with SADIE's API to update an assistant's webhook URL.

**Key Function:**
```go
func (c *Client) UpdateAssistantWebhookURL(ctx context.Context, assistantID, webhookURL string) error {
    // Makes HTTP PATCH request to /assistants/{id}
    // Updates the webhook URL for the assistant
}
```

**What it does:**
- Creates an HTTP PATCH request to SADIE's API
- Sends the new webhook URL
- SADIE updates its records
- Returns success or error

### 6. Update Webhook Handler (`internal/server/handlers/sadie_handlers.go`)

**Purpose:** This function handles requests to update the agent's webhook URL.

**Key Function:**
```go
func (h *SadieHandler) UpdateAgentWebhookHandler(c echo.Context) error {
    // 1. Determine which assistant to update
    // 2. Get the webhook URL to use
    // 3. Call SADIE API to update it
    // 4. Return success or error
}
```

**What it does:**
- Receives the HTTP request
- Figures out which assistant to update
- Gets the webhook URL (from config or request)
- Calls SADIE's API to update it
- Returns a formatted response

---

## Key Concepts Explained

### What is HTTP?

**HTTP** (HyperText Transfer Protocol) is like a language that computers use to talk to each other over the internet.

- **GET** - "Give me information" (like reading a webpage)
- **POST** - "Here's some data, do something with it" (like submitting a form)
- **PATCH** - "Update just this one thing" (like changing an address)
- **PUT** - "Update this information" (replace everything)
- **DELETE** - "Remove this information"

### What is JSON?

**JSON** (JavaScript Object Notation) is a way to format data that both humans and computers can read.

Example:
```json
{
  "name": "John",
  "age": 30,
  "phone": "+1234567890"
}
```

### What is Authentication?

**Authentication** is like showing your ID card. It proves you're allowed to access something.

In our case:
- The `x-sadie-core-secret` header is like your ID card
- The server checks if it matches
- If it matches, you're allowed in
- If it doesn't match, you get an error

### What is an API?

**API** (Application Programming Interface) is like a menu at a restaurant:
- The menu tells you what you can order
- Each item on the menu is an endpoint
- When you order something, the kitchen (server) prepares it and brings it back

---

## Common Questions

### Q: Why do we need to restart the server?

**A:** When you add new code, the server needs to reload it. Restarting the server makes it read the new code files.

### Q: What if I get an error?

**A:** Check:
1. Is the server running?
2. Is the database running?
3. Are the credentials correct?
4. Check the server logs for error messages

### Q: How do I know if it's working?

**A:** If you get a JSON response with phone numbers, it's working! If you get an error message, something needs to be fixed.

### Q: Can I use this endpoint from a web browser?

**A:** Not easily, because browsers don't let you easily add custom headers. Use curl or Postman instead.

---

## Summary

Here's what we accomplished:

1. ✅ Set up the project (database, migrations, configuration)
2. ✅ Added a function to get phone numbers from SADIE API
3. ✅ Created a handler to process HTTP requests
4. ✅ Registered a route to make the endpoint accessible
5. ✅ Tested the endpoint and got phone numbers back
6. ✅ Added webhook URL configuration (automatically determines the correct URL)
7. ✅ Created a function to update assistant webhook URL in SADIE
8. ✅ Created a handler to update the webhook URL via API
9. ✅ Registered the update endpoint
10. ✅ Tested the update functionality

**All endpoints are now live and ready to use!**

You can:
- Get phone numbers to assign to assistants
- Update the webhook URL so SADIE knows where to send events
- Call the phone numbers to test your SADIE voice assistant
- When someone calls, SADIE will use your endpoints (`/getAvailableSlips`, `/makeReservation`) to help customers
- Your server will receive webhook events at `/webhooks` when calls happen

---

## Marina Operations Endpoints

Now that we have the basic SADIE integration working, we've added Marina Operations endpoints that allow both the frontend and SADIE to interact with DockMaster API for marina-related operations.

### What are Marina Operations?

**Marina Operations** are activities related to running a marina, such as:
- Managing boat slips (parking spaces for boats)
- Handling reservations
- Creating launch tickets (for launching boats)
- Tracking launch operations

These endpoints act as a bridge between your application and DockMaster API, ensuring all marina operations go through your backend.

### The Flow

There are two main flows for these endpoints:

**Flow 1: Frontend to Backend to DockMaster**
```
dm-web-frontend → dm-web-backend → DockMaster_API
```
- The frontend makes a request to your backend
- Your backend forwards the request to DockMaster API
- DockMaster API processes it and sends back the result
- Your backend returns the result to the frontend

**Flow 2: SADIE to Backend to DockMaster**
```
SADIE (Blu) → dm-web-backend → DockMaster_API
```
- SADIE makes a request to your backend (like checking for available slips)
- Your backend forwards the request to DockMaster API
- DockMaster API processes it and sends back the result
- Your backend formats the result for SADIE and returns it

### Why This Matters

**Before these endpoints:**
- Frontend would need to call DockMaster API directly
- SADIE would need to call DockMaster API directly
- No centralized control or logging
- Harder to manage authentication and errors

**After these endpoints:**
- Everything goes through your backend
- You can log all marina operations
- You can add business logic or validation
- Easier to manage and debug
- Consistent error handling

### The Endpoints

We've added 10 Marina Operations endpoints, all under `/api/v1/MarinaOps/`:

#### 1. Get Launch Operations
**Endpoint:** `GET /api/v1/MarinaOps/LaunchOperations`

**What it does:**
- Retrieves a list of available launch operations
- Launch operations are services for launching boats (putting them in the water)

**Example use case:**
- Frontend wants to show customers what launch services are available
- SADIE needs to tell a customer what launch options they have

#### 2. Get Launch Ticket
**Endpoint:** `GET /api/v1/MarinaOps/LaunchTickets`

**What it does:**
- Retrieves information about a specific launch ticket
- A launch ticket is like a work order for launching a boat

**Example use case:**
- Customer asks: "What's the status of my launch ticket #123?"
- Frontend displays launch ticket details

#### 3. List Launch Tickets
**Endpoint:** `GET /api/v1/MarinaOps/LaunchTickets/List`

**What it does:**
- Retrieves a list of launch tickets based on search criteria
- You can filter by date, customer, status, etc.

**Example use case:**
- Show all launch tickets for today
- List all pending launch tickets

#### 4. Create Launch Ticket
**Endpoint:** `POST /api/v1/MarinaOps/LaunchTicket`

**What it does:**
- Creates a new launch ticket
- This schedules a boat to be launched

**Example use case:**
- Customer calls and wants to schedule a boat launch
- Frontend creates a launch ticket when customer books online

#### 5. Get Reservation
**Endpoint:** `GET /api/v1/MarinaOps/Reservations`

**What it does:**
- Retrieves information about a specific reservation
- A reservation is like booking a slip (parking space) for a boat

**Example use case:**
- Customer asks: "What are the details of my reservation?"
- Frontend shows reservation information

#### 6. Create or Update Reservation
**Endpoint:** `POST /api/v1/MarinaOps/Reservations`

**What it does:**
- Creates a new reservation OR updates an existing one
- This books a slip for a boat for specific dates

**Example use case:**
- SADIE helps a customer book a slip: "I'll create a reservation for you"
- Frontend allows customer to book a slip online

#### 7. Cancel Reservation
**Endpoint:** `PUT /api/v1/MarinaOps/Reservations/Cancel`

**What it does:**
- Cancels an existing reservation
- Frees up the slip for other customers

**Example use case:**
- Customer calls: "I need to cancel my reservation"
- SADIE or frontend cancels the reservation

#### 8. Get Available Slips
**Endpoint:** `GET /api/v1/MarinaOps/Slips/GetAvailableSlips`

**What it does:**
- Gets a list of slips that are available for a date range
- Only shows slips that are available for the ENTIRE date range

**Example use case:**
- Customer calls: "Do you have any slips available from March 1st to March 15th?"
- SADIE calls this endpoint to check availability
- Frontend shows available slips when customer searches

#### 9. List Slips
**Endpoint:** `GET /api/v1/MarinaOps/Slips/List`

**What it does:**
- Retrieves a list of all slips
- Shows all slips regardless of availability

**Example use case:**
- Frontend wants to show all slips in the marina
- Admin wants to see all slip information

#### 10. Search Slips
**Endpoint:** `GET /api/v1/MarinaOps/Slips/Search`

**What it does:**
- Searches for slips based on criteria
- You can search by slip number, location, size, etc.

**Example use case:**
- Customer: "Do you have a slip near the dock?"
- Frontend: "Show me slips that can fit a 40-foot boat"

### How SADIE Uses These Endpoints

When a customer calls SADIE (Blu), here's what happens:

**Scenario 1: Customer wants to check availability**
1. Customer: "Do you have any slips available next week?"
2. SADIE calls: `GET /getAvailableSlips` (SADIE endpoint)
3. Your backend calls: `GET /api/v1/MarinaOps/Slips/GetAvailableSlips` (Marina Operations endpoint)
4. Marina Operations endpoint calls DockMaster API
5. DockMaster API returns available slips
6. Your backend formats the response for SADIE
7. SADIE tells the customer: "Yes, we have 3 slips available!"

**Scenario 2: Customer wants to make a reservation**
1. Customer: "I'd like to book a slip for next week"
2. SADIE calls: `POST /makeReservation` (SADIE endpoint)
3. Your backend calls: `POST /api/v1/MarinaOps/Reservations` (Marina Operations endpoint)
4. Marina Operations endpoint calls DockMaster API
5. DockMaster API creates the reservation
6. Your backend formats the response for SADIE
7. SADIE tells the customer: "Your reservation has been created! A confirmation will be sent to you."

### Authentication

**For Frontend:**
- These endpoints require JWT authentication (user must be logged in)
- The backend automatically gets the organization ID and system ID from the logged-in user
- No need to pass organization ID or system ID manually

**For SADIE:**
- SADIE endpoints use the `x-sadie-core-secret` header for authentication
- SADIE must provide `organizationId` and `systemId` as query parameters
- This is because SADIE doesn't have a logged-in user context

### Testing the Endpoints

**Testing from Frontend (requires authentication):**

```bash
# Get available slips
curl -X GET "http://localhost:8080/api/v1/MarinaOps/Slips/GetAvailableSlips?startDate=2024-03-01&endDate=2024-03-15" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Testing from SADIE (requires SADIE secret):**

```bash
# Get available slips (SADIE endpoint)
curl -X GET "http://localhost:8080/getAvailableSlips?organizationId=YOUR_ORG_ID&systemId=YOUR_SYSTEM_ID&startDate=2024-03-01&endDate=2024-03-15" \
  -H "x-sadie-core-secret: dev_dockmaster_0195c925-ff69-7298-81a7-80f9348b52e6"
```

### What Happens Behind the Scenes

When you call a Marina Operations endpoint:

1. **Request arrives at your backend**
   - The endpoint receives the request
   - If it's from frontend, it extracts user info from JWT token
   - If it's from SADIE, it gets organization ID and system ID from query params

2. **Backend gets organization and system information**
   - Looks up the user's marina
   - Gets the organization ID and system ID for that marina
   - These are needed to call DockMaster API

3. **Backend calls DockMaster API**
   - Makes the actual API call to DockMaster
   - Includes authentication tokens
   - Passes along any query parameters or request body

4. **DockMaster API processes the request**
   - Checks availability
   - Creates reservations
   - Retrieves data
   - Returns the result

5. **Backend receives and forwards the response**
   - Gets the response from DockMaster API
   - Returns it to the caller (frontend or SADIE)
   - If it's SADIE, formats it in SADIE's expected format

### Visual Flow Diagram

```
┌─────────────┐
│   Frontend  │
│   or SADIE  │
└──────┬──────┘
       │ 1. Makes request
       │    (with auth)
       ▼
┌──────────────────┐
│  dm-web-backend  │
│  Marina Ops      │
│  Endpoint        │
└──────┬───────────┘
       │ 2. Extracts org/system ID
       │ 3. Calls DockMaster API
       ▼
┌──────────────────┐
│  DockMaster API  │
│  (External)      │
└──────┬───────────┘
       │ 4. Processes request
       │ 5. Returns result
       ▼
┌──────────────────┐
│  dm-web-backend  │
│  Formats &       │
│  Returns         │
└──────┬───────────┘
       │ 6. Sends response
       ▼
┌─────────────┐
│   Frontend  │
│   or SADIE  │
│  Receives   │
│  Response  │
└─────────────┘
```

### Files Added/Modified

**New Files:**
1. `internal/server/handlers/marina_ops_handlers.go` - All 10 endpoint handlers
2. `internal/server/routes/marina_ops_routes.go` - Route registration

**Modified Files:**
3. `internal/server/routes/routes.go` - Added Marina Operations routes registration
4. `internal/server/handlers/sadie_handlers.go` - Updated to call DockMaster API through backend

### Key Benefits

✅ **Centralized Control:** All marina operations go through your backend
✅ **Consistent Logging:** You can log all marina operations in one place
✅ **Error Handling:** Consistent error handling across all endpoints
✅ **Security:** Authentication and authorization handled in one place
✅ **Flexibility:** Easy to add business logic or validation before calling DockMaster API

## Next Steps

1. **Update the webhook URL:** Call `/updateAgentWebhook` to tell SADIE where to send webhook events
2. **Test the phone numbers:** Call one of the numbers to test your SADIE agent
3. **Monitor webhooks:** Watch your server logs to see webhook events when calls happen
4. **Test Marina Operations endpoints:** Try the endpoints from both frontend and SADIE
5. **Test SADIE integration:** Make a call and have SADIE check availability or create reservations
6. **Read the SADIE API docs:** Visit https://core-api.heysadie.ai/swagger for more information

**Important:** Make sure to update the webhook URL before testing phone calls, otherwise your server won't receive webhook events!

---

## Files Modified

### Initial Setup
1. `.env` - Created environment configuration file
2. `internal/server/handlers/generic_handlers.go` - Added root handler
3. `internal/server/routes/routes.go` - Added root route
4. `internal/server/middleware/sadie_auth.go` - Fixed unused import

### Phone Numbers Endpoint
5. `pkg/sadie/client.go` - Added `GetPhoneNumbers()` function
6. `internal/server/handlers/sadie_handlers.go` - Added `GetPhoneNumbersHandler()` function
7. `internal/server/routes/sadie_routes.go` - Added route registration for `/getPhoneNumbers`

### Webhook URL Update (Task 2)
8. `internal/config/sadie.go` - Added `WebhookURL` field and automatic URL construction
9. `pkg/sadie/client.go` - Added `UpdateAssistantWebhookURL()` function
10. `internal/server/handlers/sadie_handlers.go` - Added `UpdateAgentWebhookHandler()` function
11. `internal/server/routes/sadie_routes.go` - Added route registration for `/updateAgentWebhook`

### Marina Operations Endpoints (Task 3)
12. `internal/server/handlers/marina_ops_handlers.go` - Added all 10 Marina Operations endpoint handlers
13. `internal/server/routes/marina_ops_routes.go` - Added route registration for all Marina Operations endpoints
14. `internal/server/routes/routes.go` - Registered Marina Operations routes in main router
15. `internal/server/handlers/sadie_handlers.go` - Updated SADIE handlers to call DockMaster API through backend

---

*This guide was created to help you understand the SADIE integration process. If you have questions, refer to the code comments or the SADIE API documentation.*

