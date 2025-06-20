# Onboarding Script

This script automates the onboarding process for new organizations and users in the marina management system.

## What it does

For each row in the CSV file, the script will:

1. **Create Organization** - Creates a new organization with the specified name using the user email
2. **Create Marina** - Creates a marina under the organization with default settings
3. **Create DME Credentials** - Sets up DME API credentials for the organization
4. **Create System ID** - Creates a system ID entry for DME integration
5. **Link System ID to Marina** - Associates the system ID with the marina
6. **Create User with Invitation** - Creates a user with `org_admin` role and generates an invitation token

## CSV Format

The CSV file must have the following columns in this exact order:

| Column | Description | Required | Example |
|--------|-------------|----------|---------|
| `name` | Organization/Marina name | Yes | "Marina Corp" |
| `email` | User's email (also used for org/marina) | Yes | "john@marina.com" |
| `api_email` | DME API email/username | Yes | "api@marina.com" |
| `api_password` | DME API password | Yes | "secure_password123" |
| `system_id` | DME System ID | Yes | "SYS001" |
| `first_name` | User's first name | No | "John" |
| `last_name` | User's last name | No | "Doe" |
| `phone_number` | User's phone number | No | "555-123-4567" |
| `tier_customer_vessels` | Customer vessels tier | No (ignored) | "Free - $0" |
| `tier_customer_portal` | Customer portal tier | No (ignored) | "Free - $0" |
| `status` | Status field | No (ignored) | "" |

**Optional Field Behavior:**
- If `first_name` is empty or missing, defaults to "Admin"
- If `last_name` is empty or missing, defaults to "User"
- If `phone_number` is empty or missing, it's left blank
- Tier and status columns are ignored completely

## CSV File Location

The script looks for the CSV file at: `scripts/onboarding/onboarding_data2.csv`

## Running the Script

```bash
cd /path/to/dm-web-backend
go run scripts/onboarding/main.go
```

## User Creation

Users are created using the invitation flow:
- No password is set initially
- An invitation token is generated (valid for 10 days)
- The invitation token is logged to the console
- Users must use the invitation link to set their password

## Role Assignment

All users are created with the `org_admin` role, which provides:
- Organization management permissions
- Marina management permissions 
- User management permissions
- Limited system administration access

## Error Handling

- If any entity already exists (organization, marina, user, etc.), the script will skip creation and continue
- Errors are logged but don't stop processing of other records
- Each record is processed independently

## Output

The script provides detailed logging including:
- Which entities were created vs already existed
- Invitation tokens for new users
- Any errors encountered during processing

## Example CSV

```csv
name,email,api_email,api_password,system_id,first_name,last_name,phone_number,tier_customer_vessels,tier_customer_portal,status
Marina Corp,john@marina.com,api@marina.com,password123,SYS001,John,Doe,555-123-4567,Free - $0,Free - $0,
Harbor Inc,jane@harbor.com,api@harbor.com,password456,SYS002,Jane,Smith,555-987-6543,Free - $0,Free - $0,
Boat Services,,api@boatservices.com,password789,SYS003,,,,,Free - $0,Free - $0,
```

In the third example above, the `first_name` and `last_name` are empty, so the script will use "Admin" and "User" as defaults. 