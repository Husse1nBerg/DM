import http from "k6/http";
import { check, sleep } from "k6";
import { Rate } from "k6/metrics";

// Custom metrics - separate rates for each endpoint group
const errorRate = new Rate("error_rate");
const healthErrorRate = new Rate("health_error_rate");
const userErrorRate = new Rate("user_error_rate");
const organizationErrorRate = new Rate("organization_error_rate");
const marinaErrorRate = new Rate("marina_error_rate");
const roleErrorRate = new Rate("role_error_rate");
const addressErrorRate = new Rate("address_error_rate");
const dmeErrorRate = new Rate("dme_error_rate");
const galleryErrorRate = new Rate("gallery_error_rate");
const documentErrorRate = new Rate("document_error_rate");
const messageErrorRate = new Rate("message_error_rate");
const marinaUsageHistoryErrorRate = new Rate("marina_usage_history_error_rate");
const planErrorRate = new Rate("plan_error_rate");
const contactErrorRate = new Rate("contact_error_rate");
const permissionErrorRate = new Rate("permission_error_rate");

// Test configuration
export const options = {
  // Staging - Ramp up load testing
  stages: [
    { duration: "1m", target: 4000 }, // ramp up to 5 users over 1 minute
    { duration: "4m", target: 4000 }, // ramp up to 5 users over 1 minute
    { duration: "1m", target: 0 }, // ramp down to 0 users over 1 minute
    // { duration: "2m", target: 500 }, // Stay at 10 users for 2 minutes
    // { duration: "5m", target: 2000 }, // Ramp up to 20 users over 5 minutes
    // { duration: "5m", target: 4000 }, // Stay at 20 users for 5 minutes
    // { duration: "1m", target: 0 }, // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ["p(95)<2000"], // 95% of requests should be below 2s
    http_req_failed: ["rate<0.1"], // Error rate should be less than 10%
    error_rate: ["rate<0.1"],
    health_error_rate: ["rate<0.05"],
    user_error_rate: ["rate<0.1"],
    organization_error_rate: ["rate<0.1"],
    marina_error_rate: ["rate<0.1"],
    role_error_rate: ["rate<0.1"],
    address_error_rate: ["rate<0.1"],
    dme_error_rate: ["rate<0.15"], // DME might be less reliable
    gallery_error_rate: ["rate<0.1"],
    document_error_rate: ["rate<0.1"],
    message_error_rate: ["rate<0.1"],
    marina_usage_history_error_rate: ["rate<0.1"],
    plan_error_rate: ["rate<0.1"],
    contact_error_rate: ["rate<0.1"],
    permission_error_rate: ["rate<0.1"],
  },
};

// Configuration - Update these values for your environment
const BASE_URL = __ENV.DM_BASE_URL || "https://dmwebapi-dev.dockmaster.com";
const API_VERSION = "/api/v1";
const BASE_API_URL = `${BASE_URL}${API_VERSION}`;

// Test credentials - Update these for your test environment
const TEST_CREDENTIALS = {
  email: __ENV.DM_TEST_EMAIL || "andrew.sameh@dockmaster.com",
  password: __ENV.DM_TEST_PASSWORD || "Password@123",
};

// Global variables to store authentication tokens and test data
let authToken = "";
let refreshToken = "";
let testUserData = {};

// Global error collection
let errorLog = [];

// Helper function to track errors for both global and specific metrics
function trackError(
  success,
  specificErrorRate,
  endpoint,
  response,
  errorDetails
) {
  errorRate.add(!success);
  specificErrorRate.add(!success);

  // Log detailed error information if request failed
  if (!success) {
    const errorEntry = {
      timestamp: new Date().toISOString(),
      endpoint: endpoint || "Unknown endpoint",
      status: response ? response.status : "N/A",
      error: errorDetails || "Request failed",
      response_time: response ? response.timings.duration : "N/A",
      response_body: response
        ? response.body
          ? response.body.substring(0, 200)
          : "No body"
        : "N/A",
    };
    errorLog.push(errorEntry);

    // Also log to console for immediate visibility during test
    console.error(
      `ERROR: ${endpoint} - Status: ${errorEntry.status} - ${errorEntry.error}`
    );
    if (response && response.body) {
      console.error(`Response: ${response.body.substring(0, 200)}`);
    }
  }
}

// Simple error tracking function for backward compatibility
function trackErrorSimple(success, specificErrorRate) {
  errorRate.add(!success);
  specificErrorRate.add(!success);

  if (!success) {
    const errorEntry = {
      timestamp: new Date().toISOString(),
      endpoint: "Endpoint not specified",
      status: "Unknown",
      error: "Request failed - details not provided",
      response_time: "N/A",
      response_body: "N/A",
    };
    errorLog.push(errorEntry);
  }
}

export function setup() {
  // Authenticate and get initial test data
  console.log("Setting up test environment...");
  console.log("BASE_API_URL", BASE_API_URL);
  console.log("TEST_CREDENTIALS", TEST_CREDENTIALS);
  // Login to get authentication token
  const loginResponse = http.post(
    `${BASE_API_URL}/auth/login`,
    JSON.stringify(TEST_CREDENTIALS),
    {
      headers: { "Content-Type": "application/json" },
    }
  );

  if (loginResponse.status === 200) {
    const loginData = loginResponse.json();
    authToken = loginData.data.accessToken; // Fixed: accessToken (camelCase)
    refreshToken = loginData.data.refreshToken; // Fixed: refreshToken (camelCase)
    testUserData = loginData.data.user;
    console.log("Authentication successful");
    console.log("Token length:", authToken ? authToken.length : 0);
  } else {
    console.error(
      "Failed to authenticate:",
      loginResponse.status,
      loginResponse.body
    );
    throw new Error("Authentication failed during setup");
  }

  return {
    authToken,
    refreshToken,
    testUserData,
  };
}

export default function (data) {
  // Debug: Check if data and authToken are available
  if (!data || !data.authToken) {
    console.error("ERROR: No authentication token available in data:", data);
  }

  const headers = {
    Authorization: `Bearer ${data.authToken}`,
    "Content-Type": "application/json",
  };

  // Health check endpoint
  testHealthEndpoint();

  // User-related read endpoints
  testUserEndpoints(headers);

  // Organization read endpoints
  testOrganizationEndpoints(headers);

  // Marina read endpoints
  testMarinaEndpoints(headers);

  // Role read endpoints
  testRoleEndpoints(headers);

  // Address read endpoints
  testAddressEndpoints(headers);

  // DME read endpoints
  testDMEEndpoints(headers);

  // Gallery read endpoints
  testGalleryEndpoints(headers);

  // Document read endpoints
  testDocumentEndpoints(headers);

  // Message read endpoints
  testMessageEndpoints(headers);

  // Marina Usage History read endpoints
  testMarinaUsageHistoryEndpoints(headers);

  // Plan read endpoints
  testPlanEndpoints(headers);

  // Contact read endpoints
  testContactEndpoints(headers);

  // Permission test endpoints
  testPermissionEndpoints(headers);

  // Random sleep between 1-3 seconds to simulate real user behavior
  sleep(Math.random() * 2 + 1);
}

function testHealthEndpoint() {
  const response = http.get(`${BASE_API_URL}/health`);

  const success = check(response, {
    "health check status is 200": (r) => r.status === 200,
    "health check response time < 500ms": (r) => r.timings.duration < 500,
  });

  trackError(
    success,
    healthErrorRate,
    "/health",
    response,
    "Health check failed"
  );
}

function testUserEndpoints(headers) {
  // Get user profile
  let response = http.get(`${BASE_API_URL}/user/profile`, { headers });
  let success = check(response, {
    "user profile status is 200": (r) => r.status === 200,
    "user profile response time < 1000ms": (r) => r.timings.duration < 1000,
  });
  trackError(
    success,
    userErrorRate,
    "/user/profile",
    response,
    "User profile request failed"
  );

  // List users - using correct parameter names
  response = http.get(`${BASE_API_URL}/user/list?page=1&pageSize=10`, {
    headers,
  });
  success = check(response, {
    "user list status is 200": (r) => r.status === 200,
    "user list response time < 1500ms": (r) => r.timings.duration < 1500,
  });
  trackError(
    success,
    userErrorRate,
    "/user/list",
    response,
    "User list request failed"
  );

  // Get users by role - first get role list to get a real role ID
  const rolesResponse = http.get(`${BASE_API_URL}/role/list`, { headers });
  if (rolesResponse.status === 200) {
    const rolesData = rolesResponse.json();
    if (rolesData.data && rolesData.data.length > 0) {
      const roleId = rolesData.data[0].id;
      response = http.get(`${BASE_API_URL}/user/role/${roleId}`, { headers });
      success = check(response, {
        "users by role response is valid": (r) =>
          r.status === 200 || r.status === 404,
      });
      trackError(
        response.status === 200 || response.status === 404,
        userErrorRate,
        `/user/role/${roleId}`,
        response,
        "Get users by role request failed"
      );
    }
  }
}

function testOrganizationEndpoints(headers) {
  // Get organizations paginated - using correct parameter names
  let response = http.get(`${BASE_API_URL}/organizations?page=1&pageSize=10`, {
    headers,
  });
  let success = check(response, {
    "organizations list status is 200": (r) => r.status === 200,
    "organizations list response time < 1500ms": (r) =>
      r.timings.duration < 1500,
  });
  trackError(
    success,
    organizationErrorRate,
    "/organizations",
    response,
    "Organizations list request failed"
  );

  // Get organization by email - use first organization's email from list if available
  if (response.status === 200) {
    const orgData = response.json();
    if (orgData.data && orgData.data.length > 0 && orgData.data[0].email) {
      response = http.get(
        `${BASE_API_URL}/organizations/by-email?email=${orgData.data[0].email}`,
        { headers }
      );
      success = check(response, {
        "organization by email response is valid": (r) =>
          r.status === 200 || r.status === 404,
      });
      trackError(
        response.status === 200 || response.status === 404,
        organizationErrorRate,
        "/organizations/by-email",
        response,
        "Organization by email request failed"
      );
    }
  }
}

function testMarinaEndpoints(headers) {
  // Get marinas paginated - using correct parameter names
  let response = http.get(`${BASE_API_URL}/marinas?page=1&pageSize=10`, {
    headers,
  });
  let success = check(response, {
    "marinas list status is 200": (r) => r.status === 200,
    "marinas list response time < 1500ms": (r) => r.timings.duration < 1500,
  });
  trackError(
    success,
    marinaErrorRate,
    "/marinas",
    response,
    "Marinas list request failed"
  );

  // Get user marinas
  response = http.get(`${BASE_API_URL}/marinas/user`, { headers });
  success = check(response, {
    "user marinas response is valid": (r) =>
      r.status === 200 || r.status === 404,
  });
  trackError(
    response.status === 200 || response.status === 404,
    marinaErrorRate,
    "/marinas/user",
    response,
    "User marinas request failed"
  );

  // Get marina by email - use first marina's email from list if available
  const marinasListResponse = http.get(
    `${BASE_API_URL}/marinas?page=1&pageSize=10`,
    { headers }
  );
  if (marinasListResponse.status === 200) {
    const marinasListData = marinasListResponse.json();
    if (
      marinasListData.data &&
      marinasListData.data.length > 0 &&
      marinasListData.data[0].email
    ) {
      response = http.get(
        `${BASE_API_URL}/marinas/by-email?email=${marinasListData.data[0].email}`,
        { headers }
      );
      success = check(response, {
        "marina by email response is valid": (r) =>
          r.status === 200 || r.status === 404,
      });
      trackError(
        response.status === 200 || response.status === 404,
        marinaErrorRate,
        "/marinas/by-email",
        response,
        "Marina by email request failed"
      );
    }
  }
}

function testRoleEndpoints(headers) {
  // List roles
  let response = http.get(`${BASE_API_URL}/role/list`, { headers });
  let success = check(response, {
    "roles list status is 200": (r) => r.status === 200,
    "roles list response time < 1000ms": (r) => r.timings.duration < 1000,
  });
  trackError(
    success,
    roleErrorRate,
    "/role/list",
    response,
    "Roles list request failed"
  );

  // Get role by name - use first role's name from list if available
  if (response.status === 200) {
    const rolesData = response.json();
    if (rolesData.data && rolesData.data.length > 0 && rolesData.data[0].name) {
      response = http.get(
        `${BASE_API_URL}/role/name?name=${rolesData.data[0].name}`,
        { headers }
      );
      success = check(response, {
        "role by name response is valid": (r) =>
          r.status === 200 || r.status === 404,
      });
      trackError(
        response.status === 200 || response.status === 404,
        roleErrorRate,
        "/role/name",
        response,
        "Role by name request failed"
      );
    }
  }
}

function testAddressEndpoints(headers) {
  // First get the user profile to get address IDs
  const profileResponse = http.get(`${BASE_API_URL}/user/profile`, { headers });
  if (profileResponse.status === 200) {
    const profileData = profileResponse.json();
    if (profileData.data && profileData.data.address_id) {
      // Get address by actual ID from user profile
      const response = http.get(
        `${BASE_API_URL}/addresses/${profileData.data.address_id}`,
        { headers }
      );
      const success = check(response, {
        "address by id response is valid": (r) =>
          r.status === 200 || r.status === 404,
      });
      trackError(
        response.status === 200 || response.status === 404,
        addressErrorRate,
        "/addresses",
        response,
        "Address by id request failed"
      );
    }
  }
}

function testDMEEndpoints(headers) {
  // List DME System IDs
  let response = http.get(`${BASE_API_URL}/dme/sysids`, { headers });
  let success = check(response, {
    "dme sysids list response is valid": (r) =>
      r.status === 200 || r.status === 404,
  });
  trackError(
    response.status === 200 || response.status === 404,
    dmeErrorRate,
    "/dme/sysids",
    response,
    "DME sysids list request failed"
  );

  // Get DME credential by organization - first get user's organization
  const profileResponse = http.get(`${BASE_API_URL}/user/profile`, { headers });
  if (profileResponse.status === 200) {
    const profileData = profileResponse.json();
    if (profileData.data && profileData.data.organization_id) {
      response = http.get(
        `${BASE_API_URL}/dme/credentials/organization/${profileData.data.organization_id}`,
        {
          headers,
        }
      );
      success = check(response, {
        "dme credentials response is valid": (r) =>
          r.status === 200 || r.status === 404,
      });
      trackError(
        response.status === 200 || response.status === 404,
        dmeErrorRate,
        "/dme/credentials/organization",
        response,
        "DME credentials request failed"
      );
    }
  }
}

function testGalleryEndpoints(headers) {
  // First get user marinas to get a real marina ID
  const marinasResponse = http.get(`${BASE_API_URL}/marinas/user`, { headers });
  if (marinasResponse.status === 200) {
    const marinasData = marinasResponse.json();
    if (marinasData.data && marinasData.data.length > 0) {
      const marinaId = marinasData.data[0].id;

      // Get marina gallery using real marina ID
      let response = http.get(`${BASE_API_URL}/gallery/marina/${marinaId}`, {
        headers,
      });
      let success = check(response, {
        "marina gallery response is valid": (r) =>
          r.status === 200 || r.status === 404,
      });
      trackError(
        response.status === 200 || response.status === 404,
        galleryErrorRate,
        "/gallery/marina",
        response,
        "Marina gallery request failed"
      );

      // If gallery has items, get the first one
      if (response.status === 200) {
        const galleryData = response.json();
        if (galleryData.data && galleryData.data.length > 0) {
          const itemId = galleryData.data[0].id;
          response = http.get(`${BASE_API_URL}/gallery/marina/item/${itemId}`, {
            headers,
          });
          success = check(response, {
            "marina gallery item response is valid": (r) =>
              r.status === 200 || r.status === 404,
          });
          trackError(
            response.status === 200 || response.status === 404,
            galleryErrorRate,
            "/gallery/marina/item",
            response,
            "Marina gallery item request failed"
          );
        }
      }
    }
  }
}

function testDocumentEndpoints(headers) {
  // Skip document endpoints in load test since they require marina context
  // and complex entity relationships for load testing purposes
  console.log("Skipping document endpoints - require marina context");
}

function testMessageEndpoints(headers) {
  // Skip message endpoints in load test since they require both marinaId AND customerId
  // which are complex to obtain dynamically for load testing purposes
  console.log("Skipping message endpoints - require customer context");
}

function testMarinaUsageHistoryEndpoints(headers) {
  // First get user marinas to get a real marina ID
  const marinasResponse = http.get(`${BASE_API_URL}/marinas/user`, { headers });
  if (marinasResponse.status === 200) {
    const marinasData = marinasResponse.json();
    if (marinasData.data && marinasData.data.length > 0) {
      const marinaId = marinasData.data[0].id;

      // Get marina usage history by marina ID using real marina ID
      let response = http.get(
        `${BASE_API_URL}/marina-usage-history/marina/${marinaId}`,
        {
          headers,
        }
      );
      let success = check(response, {
        "marina usage history response is valid": (r) =>
          r.status === 200 || r.status === 404,
      });
      trackError(
        response.status === 200 || response.status === 404,
        marinaUsageHistoryErrorRate,
        "/marina-usage-history/marina",
        response,
        "Marina usage history request failed"
      );

      // Get latest marina usage history
      response = http.get(
        `${BASE_API_URL}/marina-usage-history/marina/${marinaId}/latest`,
        {
          headers,
        }
      );
      success = check(response, {
        "latest marina usage history response is valid": (r) =>
          r.status === 200 || r.status === 404,
      });
      trackError(
        response.status === 200 || response.status === 404,
        marinaUsageHistoryErrorRate,
        "/marina-usage-history/marina/latest",
        response,
        "Latest marina usage history request failed"
      );
    }
  }
}

function testPlanEndpoints(headers) {
  // List notes and messages plans with explicit pagination to avoid negative offset
  let response = http.get(
    `${BASE_API_URL}/plans/notes-messages?page=1&pageSize=10`,
    { headers }
  );
  let success = check(response, {
    "notes messages plans response is valid": (r) =>
      r.status === 200 || r.status === 404,
  });
  trackError(
    response.status === 200 || response.status === 404,
    planErrorRate,
    "/plans/notes-messages",
    response,
    "Notes messages plans request failed"
  );

  // List storage plans with explicit pagination to avoid negative offset
  response = http.get(`${BASE_API_URL}/plans/storage?page=1&pageSize=10`, {
    headers,
  });
  success = check(response, {
    "storage plans response is valid": (r) =>
      r.status === 200 || r.status === 404,
  });
  trackError(
    response.status === 200 || response.status === 404,
    planErrorRate,
    "/plans/storage",
    response,
    "Storage plans request failed"
  );
}

function testContactEndpoints(headers) {
  // First get user marinas to get a real marina ID
  const marinasResponse = http.get(`${BASE_API_URL}/marinas/user`, { headers });
  if (marinasResponse.status === 200) {
    const marinasData = marinasResponse.json();
    if (marinasData.data && marinasData.data.length > 0) {
      const marinaId = marinasData.data[0].id;

      // Get contacts for marina using real marina ID
      const response = http.get(
        `${BASE_API_URL}/marinas/${marinaId}/contacts`,
        { headers }
      );
      const success = check(response, {
        "marina contacts response is valid": (r) =>
          r.status === 200 || r.status === 404,
      });
      trackError(
        response.status === 200 || response.status === 404,
        contactErrorRate,
        "/marinas/contacts",
        response,
        "Marina contacts request failed"
      );
    }
  }
}

function testPermissionEndpoints(headers) {
  // Get user permissions
  let response = http.get(`${BASE_API_URL}/test/permissions/user`, { headers });
  let success = check(response, {
    "user permissions response is valid": (r) =>
      r.status === 200 || r.status === 404,
  });
  trackError(
    response.status === 200 || response.status === 404,
    permissionErrorRate,
    "/test/permissions/user",
    response,
    "User permissions request failed"
  );

  // Get route permissions
  response = http.get(`${BASE_API_URL}/test/permissions/routes`, { headers });
  success = check(response, {
    "route permissions response is valid": (r) =>
      r.status === 200 || r.status === 404,
  });
  trackError(
    response.status === 200 || response.status === 404,
    permissionErrorRate,
    "/test/permissions/routes",
    response,
    "Route permissions request failed"
  );
}

export function teardown(data) {
  console.log("Test completed");
  console.log(
    "Authentication token used:",
    data && data.authToken ? "Yes" : "No"
  );

  // Output error log summary
  console.log("\n=== ERROR LOG SUMMARY ===");
  console.log(`Total errors recorded: ${errorLog.length}`);

  if (errorLog.length > 0) {
    // Group errors by endpoint for summary
    const errorsByEndpoint = {};
    errorLog.forEach((error) => {
      if (!errorsByEndpoint[error.endpoint]) {
        errorsByEndpoint[error.endpoint] = [];
      }
      errorsByEndpoint[error.endpoint].push(error);
    });

    console.log("\nErrors by endpoint:");
    Object.keys(errorsByEndpoint).forEach((endpoint) => {
      console.log(`  ${endpoint}: ${errorsByEndpoint[endpoint].length} errors`);
    });

    // Output detailed error log in JSON format for file saving
    console.log("\n=== DETAILED ERROR LOG (JSON) ===");
    console.log("ERROR_LOG_START");
    console.log(
      JSON.stringify(
        {
          test_completed_at: new Date().toISOString(),
          total_errors: errorLog.length,
          errors: errorLog,
        },
        null,
        2
      )
    );
    console.log("ERROR_LOG_END");

    // Instructions for saving to file
    console.log("\n=== SAVE ERROR LOG TO FILE ===");
    console.log(
      "To save the error log to a file, run k6 with output redirection:"
    );
    console.log("k6 run k6-load-test.js > test-results.log 2>&1");
    console.log(
      "Then extract the JSON between ERROR_LOG_START and ERROR_LOG_END"
    );
    console.log("Or use this command to extract just the error log:");
    console.log(
      "k6 run k6-load-test.js 2>&1 | sed -n '/ERROR_LOG_START/,/ERROR_LOG_END/p' | sed '1d;$d' > error-log.json"
    );
  } else {
    console.log("No errors recorded during the test! 🎉");
  }
}
