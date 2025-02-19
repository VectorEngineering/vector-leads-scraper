# Vector Leads Scraper API Specification

## Overview
This document provides detailed specifications for the Vector Leads Scraper service APIs, focusing on new analytics, reporting, and enhanced search capabilities.

## API Versioning
All new APIs will be versioned using protobuf package versioning:
```protobuf
package vector.leadsscraper.v1;
```

## Common Types

### Status Codes
```protobuf
enum StatusCode {
  STATUS_CODE_UNSPECIFIED = 0;
  STATUS_CODE_SUCCESS = 1;
  STATUS_CODE_INVALID_REQUEST = 2;
  STATUS_CODE_UNAUTHORIZED = 3;
  STATUS_CODE_FORBIDDEN = 4;
  STATUS_CODE_NOT_FOUND = 5;
  STATUS_CODE_INTERNAL_ERROR = 6;
}
```

### Common Response
```protobuf
message Response {
  StatusCode status = 1;
  string message = 2;
  google.protobuf.Any data = 3;
}
```

## Analytics Service

### Job Metrics API
```protobuf
message JobMetricsRequest {
  string organization_id = 1;
  string workspace_id = 2;
  TimeRange time_range = 3;
  repeated string metric_types = 4;
  map<string, string> filters = 5;
}

message JobMetricsResponse {
  message Metric {
    string name = 1;
    double value = 2;
    string unit = 3;
    google.protobuf.Timestamp timestamp = 4;
  }
  
  repeated Metric metrics = 1;
  map<string, string> aggregations = 2;
}

rpc GetJobMetrics(JobMetricsRequest) returns (JobMetricsResponse);
```

### Lead Insights API
```protobuf
message LeadInsightsRequest {
  string organization_id = 1;
  string workspace_id = 2;
  TimeRange time_range = 3;
  repeated string dimensions = 4;
  repeated string metrics = 5;
}

message LeadInsightsResponse {
  message Insight {
    string dimension = 1;
    map<string, double> metrics = 2;
    repeated string segments = 3;
  }
  
  repeated Insight insights = 1;
  map<string, string> metadata = 2;
}

rpc GetLeadInsights(LeadInsightsRequest) returns (LeadInsightsResponse);
```

### System Health API
```protobuf
message SystemHealthRequest {
  string organization_id = 1;
  repeated string components = 2;
}

message SystemHealthResponse {
  message ComponentHealth {
    string name = 1;
    string status = 2;
    map<string, string> metrics = 3;
    repeated string alerts = 4;
  }
  
  repeated ComponentHealth components = 1;
  google.protobuf.Timestamp timestamp = 2;
}

rpc GetSystemHealth(SystemHealthRequest) returns (SystemHealthResponse);
```

## Search Service

### Advanced Search API
```protobuf
message AdvancedSearchRequest {
  string organization_id = 1;
  string workspace_id = 2;
  SearchCriteria criteria = 3;
  Pagination pagination = 4;
  Sorting sorting = 5;
}

message SearchCriteria {
  repeated Filter filters = 1;
  repeated string fields = 2;
  string query = 3;
}

message Filter {
  string field = 1;
  string operator = 2;
  google.protobuf.Any value = 3;
}

message AdvancedSearchResponse {
  repeated Lead leads = 1;
  SearchMetadata metadata = 2;
  Pagination pagination = 3;
}

rpc AdvancedSearch(AdvancedSearchRequest) returns (AdvancedSearchResponse);
```

### Lead Filtering API
```protobuf
message FilterLeadsRequest {
  string organization_id = 1;
  string workspace_id = 2;
  repeated FilterGroup filter_groups = 3;
  Pagination pagination = 4;
}

message FilterGroup {
  repeated Filter filters = 1;
  string operator = 2;  // AND, OR
}

message FilterLeadsResponse {
  repeated Lead leads = 1;
  FilterMetadata metadata = 2;
  Pagination pagination = 3;
}

rpc FilterLeads(FilterLeadsRequest) returns (FilterLeadsResponse);
```

### Search Aggregations API
```protobuf
message SearchAggregationsRequest {
  string organization_id = 1;
  string workspace_id = 2;
  repeated string aggregation_fields = 3;
  SearchCriteria criteria = 4;
}

message SearchAggregationsResponse {
  message Aggregation {
    string field = 1;
    map<string, int64> buckets = 2;
    Statistics statistics = 3;
  }
  
  repeated Aggregation aggregations = 1;
  SearchMetadata metadata = 2;
}

rpc SearchAggregations(SearchAggregationsRequest) returns (SearchAggregationsResponse);
```

## Error Handling

### Error Response
```protobuf
message Error {
  string code = 1;
  string message = 2;
  map<string, string> details = 3;
  string tracking_id = 4;
}
```

### Common Error Codes
1. `INVALID_REQUEST`: Request validation failed
2. `UNAUTHORIZED`: Authentication required
3. `FORBIDDEN`: Insufficient permissions
4. `NOT_FOUND`: Resource not found
5. `INTERNAL`: Internal server error
6. `RATE_LIMITED`: Too many requests

## Authentication & Authorization

### Required Headers
1. `Authorization`: Bearer token or API key
2. `X-Organization-ID`: Organization identifier
3. `X-Correlation-ID`: Request correlation ID

### Scope Requirements
Each endpoint requires specific scopes:
1. Analytics APIs: `analytics:read`
2. Search APIs: `leads:read`
3. System Health: `system:read`

## Rate Limiting
- Default: 1000 requests per minute per organization
- Burst: 100 requests per second
- Headers:
  * `X-RateLimit-Limit`
  * `X-RateLimit-Remaining`
  * `X-RateLimit-Reset`

## Pagination
```protobuf
message Pagination {
  int32 page_size = 1;
  string page_token = 2;
  int32 total_items = 3;
  bool has_more = 4;
}
```

## Implementation Guidelines

### Error Handling
1. Use appropriate status codes
2. Include detailed error messages
3. Add request tracking IDs
4. Log all errors with context

### Performance
1. Use cursor-based pagination
2. Implement request timeouts
3. Cache frequently accessed data
4. Use appropriate indexes

### Security
1. Validate all input
2. Sanitize output
3. Implement rate limiting
4. Use proper authentication

### Testing
1. Write unit tests for all endpoints
2. Include integration tests
3. Perform load testing
4. Document test cases 