#!/bin/bash

# Exit on error, but allow for cleanup
set -e

# Cleanup function
cleanup() {
    local exit_code=$?
    echo "Performing cleanup..."
    # Kill any remaining grpcurl processes
    pkill -f grpcurl || true
    exit $exit_code
}

# Set up cleanup trap
trap cleanup EXIT

# Install grpcurl if not already installed
if ! command -v grpcurl &> /dev/null; then
    echo "Installing grpcurl..."
    go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
fi

# Set path to grpcurl
GRPCURL="$HOME/go/bin/grpcurl"

# Function to make gRPC call and handle response
make_grpc_call() {
    local name="$1"
    local service="$2"
    local request="$3"
    local max_retries=3
    local retry_count=0
    local wait_time=2
    
    while [ $retry_count -lt $max_retries ]; do
        echo "Making gRPC call to $service/$name (attempt $((retry_count + 1))/$max_retries)"
        echo "Request:"
        echo "$request" | jq . || echo "$request"
        
        # Make the gRPC call with verbose output and capture both stdout and stderr
        local output
        if output=$("$GRPCURL" -v -plaintext \
            -H "x-tenant-id: 0" \
            -H "x-organization-id: 0" \
            -d "$request" localhost:50051 "$service/$name" 2>&1); then
            
            # Extract the response part after "Response contents:"
            local response
            response=$(echo "$output" | awk '/Response contents:/{p=1;next} p{print}')
            
            echo "Response:"
            if echo "$response" | jq . > /dev/null 2>&1; then
                echo "$response" | jq .
                
                # Try to parse as JSON and extract ID
                local id
                if [[ "$name" == "CreateOrganization" ]]; then
                    id=$(echo "$response" | jq -r '.organization.id // empty')
                elif [[ "$name" == "CreateTenant" ]]; then
                    id=$(echo "$response" | jq -r '.tenant.id // empty')
                elif [[ "$name" == "CreateAccount" ]]; then
                    id=$(echo "$response" | jq -r '.account.id // empty')
                fi
                
                if [[ -n "$id" ]]; then
                    echo "Extracted ID: $id"
                    echo "$id"
                    return 0
                fi
            else
                echo "Raw response: $response"
            fi
        fi
        
        echo "Call failed or invalid response, retrying in $wait_time seconds..."
        sleep $wait_time
        retry_count=$((retry_count + 1))
        wait_time=$((wait_time * 2))
    done
    
    echo "Failed to make gRPC call after $max_retries attempts"
    return 1
}

echo "Testing gRPC API endpoints..."

# List available services
echo "Listing available services..."
if ! "$GRPCURL" -plaintext \
    -H "x-tenant-id: 0" \
    -H "x-organization-id: 0" \
    localhost:50051 list; then
    echo "Failed to list services"
    exit 1
fi
echo "✓ SUCCESS: Services listed successfully"

# Set service name
service="lead_scraper_service.v1.LeadScraperService"

# Create organization
name="CreateOrganization"
request=$(cat <<EOF
{
  "organization": {
    "name": "Test Organization",
    "description": "A test organization",
    "website": "https://test.org",
    "max_tenants": 3,
    "max_api_keys": 10,
    "max_users": 100,
    "billing_email": "billing@example.com",
    "technical_email": "tech@example.com"
  }
}
EOF
)

org_id=$(make_grpc_call "$name" "$service" "$request") || exit 1
echo "Organization created with ID: $org_id"

# Create tenant
name="CreateTenant"
request=$(cat <<EOF
{
  "tenant": {
    "name": "Test Tenant",
    "description": "A test tenant",
    "organization_id": "$org_id"
  }
}
EOF
)

tenant_id=$(make_grpc_call "$name" "$service" "$request") || exit 1
echo "Tenant created with ID: $tenant_id"

# Create account
name="CreateAccount"
request=$(cat <<EOF
{
  "account": {
    "name": "Test Account",
    "description": "A test account",
    "organization_id": "$org_id",
    "tenant_id": "$tenant_id"
  }
}
EOF
)

account_id=$(make_grpc_call "$name" "$service" "$request") || exit 1
echo "Account created with ID: $account_id"

# Get account
name="GetAccount"
request=$(cat <<EOF
{
  "id": "$account_id"
}
EOF
)

make_grpc_call "$name" "$service" "$request" || exit 1

# List accounts
name="ListAccounts"
request=$(cat <<EOF
{
  "organization_id": "$org_id",
  "tenant_id": "$tenant_id"
}
EOF
)

make_grpc_call "$name" "$service" "$request" || exit 1

# Update account
name="UpdateAccount"
request=$(cat <<EOF
{
  "account": {
    "id": "$account_id",
    "name": "Updated Test Account",
    "description": "An updated test account",
    "organization_id": "$org_id",
    "tenant_id": "$tenant_id"
  }
}
EOF
)

make_grpc_call "$name" "$service" "$request" || exit 1

# Delete account
name="DeleteAccount"
request=$(cat <<EOF
{
  "id": "$account_id"
}
EOF
)

make_grpc_call "$name" "$service" "$request" || exit 1

echo "✓ All tests completed successfully" 