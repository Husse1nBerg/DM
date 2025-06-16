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

| Column | Description | Example |
|--------|-------------|---------|
| `org_name` | Organization name | "Marina Corp" |
| `marina_name` | Marina name | "Sunset Marina" |
| `dme_email` | DME API email/username | "api@marina.com" |
| `dme_password` | DME API password | "secure_password123" |
| `system_id` | DME System ID | "SYS001" |
| `user_first_name` | User's first name | "John" |
| `user_last_name` | User's last name | "Doe" |
| `user_email` | User's email (also used for org/marina) | "john@marina.com" |

## CSV File Location

The script looks for the CSV file at: `scripts/onboarding/onboarding_data.csv`

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
org_name,marina_name,dme_email,dme_password,system_id,user_first_name,user_last_name,user_email
Marina Corp,Sunset Marina,api@marina.com,password123,SYS001,John,Doe,john@marina.com
Harbor Inc,Harbor Marina,api@harbor.com,password456,SYS002,Jane,Smith,jane@harbor.com
``` 