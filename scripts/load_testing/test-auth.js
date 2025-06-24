import http from "k6/http";

const BASE_URL = __ENV.DM_BASE_URL || "https://dmwebapi-dev.dockmaster.com";
const API_VERSION = "/api/v1";
const BASE_API_URL = `${BASE_URL}${API_VERSION}`;

const TEST_CREDENTIALS = {
  email: __ENV.DM_TEST_EMAIL || "andrew.sameh@dockmaster.com",
  password: __ENV.DM_TEST_PASSWORD || "Password@123",
};

export default function () {
  console.log("Testing authentication...");
  console.log("BASE_API_URL:", BASE_API_URL);
  console.log("Credentials:", TEST_CREDENTIALS);

  // Test login
  const loginResponse = http.post(
    `${BASE_API_URL}/auth/login`,
    JSON.stringify(TEST_CREDENTIALS),
    {
      headers: { "Content-Type": "application/json" },
    }
  );

  console.log("Login response status:", loginResponse.status);
  console.log("Login response body:", loginResponse.body);

  if (loginResponse.status === 200) {
    const loginData = loginResponse.json();
    const token = loginData.data.accessToken; // Fixed: accessToken (camelCase)
    console.log("Token received:", token ? "Yes" : "No");
    console.log("Token length:", token ? token.length : 0);

    // Test authenticated request
    const headers = {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    };

    const profileResponse = http.get(`${BASE_API_URL}/user/profile`, {
      headers,
    });
    console.log("Profile response status:", profileResponse.status);
    console.log("Profile response body:", profileResponse.body);
  } else {
    console.error("Login failed!");
  }
}
