# Approval Service

## Overview

The Approval Service is a web application built using Go and the Gin framework. It provides endpoints for creating, approving, and checking the status of approvals. The service interacts with Slack for sending approval buttons and handles approval management through a custom `ApprovalManager`.

## Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/your-repo/approval-service.git
   cd approval-service
   ```

2. Install dependencies:

   ```sh
   go mod tidy
   ```

3. Build the server:
   ```sh
   go build -o approval-server
   ```

## Usage

Start the server:

```sh
./approval-server
```

By default, the server runs on `http://localhost:8080`.

## API Endpoints

### Create Approval

- **URL:** `POST /:org_id/approvals`
- **Description:** Creates a new approval request and sends an approval button to Slack.
- **Payload:** `RequestApprovalPayload` (customize as needed)

### Approve Approval

- **URL:** `POST /:org_id/approvals/:approval_id/approve`
- **Description:** Approves the specified approval request.

### Get Approval Status

- **URL:** `GET /:org_id/approvals/:approval_id`
- **Description:** Retrieves the status of the specified approval request.

### List Approvals

- **URL:** `GET /:org_id/approvals`
- **Description:** Lists all approvals for the specified organization.

## Configuration

- **SlackManager:** Manages interactions with Slack, including sending approval buttons.
- **ApprovalManager:** Manages approval requests and statuses.

# Approval Client

## Overview

The Approval Client is a Go application that interacts with the Approval Service to create an approval request and wait for it to be approved. The client periodically checks the status of the approval until it is either approved or a specified total wait time elapses.

## Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/your-repo/approval-client.git
   cd approval-client
   ```

2. Install dependencies:

   ```sh
   go mod tidy
   ```

3. Build the client:
   ```sh
   go build -o approval-client
   ```

## Usage

Run the client:

```sh
./approval-client --org_id <org_id> --wait_time <wait_time> --base_url <base_url>
```

### Flags

- **`--org_id`:** The organization ID (default: `example_org`)
- **`--wait_time`:** The total time to wait for approval (e.g., `30s`, `1m`)
- **`--base_url`:** The base URL of the approval service (default: `http://localhost:8080`)

### Example

```sh
./approval-client --org_id my_org --wait_time 15m --base_url https://approvals.fly.dev --message "*This service is about to deploy. You have 5 minutes to stop deploy if you do not want it.*'
```

## Configuration

- **`org_id`:** The organization ID used for the approval request.
- **`wait_time`:** The total time the client will wait for the approval to be approved.
- **`base_url`:** The base URL of the Approval Service.
