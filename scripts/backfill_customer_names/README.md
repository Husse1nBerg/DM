# Customer Name Backfill Script

This script backfills the `customer_name` field in the `esign_submissions` table by fetching customer data from the DME (DockMaster Enterprise) API.

## Purpose

When the `customer_name` field was added to the submissions table, existing submissions only had `customer_id` but not the actual customer name. This script fills in those missing names by:

1. Finding all submissions with a `customer_id` but no `customer_name`
2. Fetching the customer information from DME API
3. Updating the submission with the customer's name

## Usage

From the project root, run:

```bash
go run scripts/backfill_customer_names/main.go
```

## Requirements

- The application must have valid DME credentials configured
- Marina records must have valid `system_id` values
- Internet connection to reach the DME API

## Features

- **Rate Limiting**: Waits 100ms between API calls to avoid overwhelming the DME API
- **Error Handling**: Continues processing even if some records fail
- **Detailed Logging**: Shows progress and results for each submission
- **Summary Report**: Provides statistics at the end of execution

## Output

The script will display:
- Total submissions processed
- Successfully updated count
- Errors encountered
- Submissions skipped (no system ID or no customer name in DME)

## Safety

- Does not modify submissions that already have a `customer_name`
- Only updates submissions with a valid `customer_id`
- Logs warnings but continues on errors (graceful degradation)

## Exit Codes

- `0`: Success (all processable submissions were updated)
- `1`: Partial failure (some submissions had errors)



