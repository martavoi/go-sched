# Couchbase Job Scheduler Example

This example demonstrates the go-sched library with **Couchbase persistence**, showcasing enterprise-grade NoSQL storage with the same concurrent job processing and fault tolerance features.

## Prerequisites

### 1. **Couchbase Server**
You need a running Couchbase instance. Choose one:

**Option A: Docker (Recommended)**
```bash
# Run Couchbase in Docker
docker run -d --name couchbase-scheduler \
  -p 8091-8096:8091-8096 \
  -p 11210-11211:11210-11211 \
  couchbase:latest
```

**Option B: Local Couchbase**
- Download from [couchbase.com/downloads](https://www.couchbase.com/downloads)
- Install and start Couchbase Server

**Option C: Couchbase Cloud**
- Create free cluster at [cloud.couchbase.com](https://cloud.couchbase.com)
- Get connection string and credentials

### 2. **Bucket and Collection Setup**
Create the required bucket, scope, and collection in your Couchbase cluster:

**Step 1: Create Bucket**
1. Access Couchbase Web Console (http://localhost:8091)
2. Go to **Buckets** → **Add Bucket**
3. Name: `scheduler`
4. Memory Quota: 256 MB (minimum)
5. Click **Add Bucket**

**Step 2: Create Scope and Collection**
1. Go to **Scopes & Collections** → Select `scheduler` bucket
2. Click **Add Scope**
3. Name: `jobs`
4. Click **Save**
5. In the `jobs` scope, click **Add Collection**
6. Name: `email-jobs`
7. Click **Save**

**Required**: This example requires Couchbase 7.0+ with scopes and collections enabled.

### 3. **Go Dependencies**
```bash
# Install Couchbase driver (if not already installed)
go mod tidy
```

## Configuration

### **Environment Variables**
```bash
# Optional: Set Couchbase connection details
export COUCHBASE_URI="couchbase://localhost"
export COUCHBASE_USERNAME="Administrator"
export COUCHBASE_PASSWORD="password"

# For Couchbase Cloud:
export COUCHBASE_URI="couchbases://cb.your-cluster.cloud.couchbase.com"
export COUCHBASE_USERNAME="your-username"
export COUCHBASE_PASSWORD="your-password"
```

**Defaults**: 
- URI: `couchbase://localhost`
- Username: `Administrator`
- Password: `password`

## Running the Example

```bash
cd examples/couchbase
go run main.go
```

## What This Example Does

- **Creates 20 email jobs** with random scheduling (1-30 seconds delay)
- **5 concurrent workers** process emails with random durations (1-5 seconds)
- **30-second visibility timeout** for automatic fault recovery
- **Couchbase persistence** - jobs survive application restarts
- **Type-safe email processing** - demonstrates structured payloads
- **Uses modern collections** (requires Couchbase 7.0+ with `jobs` scope and `email-jobs` collection)
- **Demonstrates graceful shutdown** with immediate job visibility reset

## Key Features Demonstrated

### 1. **NoSQL Document Storage**
- Jobs stored as JSON documents in Couchbase
- Leverages Couchbase's distributed architecture
- High availability and automatic failover

### 2. **N1QL Query Integration**
- Uses N1QL (SQL for JSON) for efficient job queries
- Supports complex filtering and ordering
- Optimized for high-performance operations

### 3. **Enterprise Scalability**
- Multi-dimensional scaling (memory, compute, storage)
- Built-in caching and indexing
- Sub-millisecond data access

### 4. **Fault Tolerance**
- Jobs automatically become visible again after 30s if worker crashes
- Document persistence ensures no job loss on application restart
- Graceful shutdown makes remaining jobs immediately available

## Document Structure

The example creates documents in the `scheduler` bucket:

### **Pending Job Document**
```json
{
  "id": "4a9188cf-0d3e-45fe-9e19-1aa5e13c2745",
  "status": "pending",
  "processAfter": "2025-07-18T21:53:23.722Z",
  "payload": {
    "to": "bob@company.org",
    "subject": "notification email #1",
    "type": "notification"
  }
}
```

### **Completed Job Document**
```json
{
  "id": "a5c64868-8a15-4449-848b-2df0a71c3280",
  "status": "completed",
  "processAfter": "2025-07-18T21:53:13.733Z",
  "payload": {
    "to": "alice@example.com",
    "subject": "reminder email #2",
    "type": "reminder"
  },
  "processedAt": "2025-07-18T21:53:16.821Z",
  "visibleAfter": null
}
```

### **Field Descriptions**
| Field | Type | Description |
|-------|------|-------------|
| `id` | String | Unique job identifier (UUID) |
| `status` | String | Job state: `"pending"` or `"completed"` |
| `processAfter` | Date | When job should be processed |
| `visibleAfter` | Date/null | Visibility timeout (null = visible) |
| `processedAt` | Date/null | When job completed (only for completed jobs) |
| `payload` | Object | EmailJob data with `to`, `subject`, `type` |

## Sample Execution Log

```
time=2025-07-19T01:15:01.721+03:00 level=INFO msg="connected to Couchbase" uri=couchbase://localhost
time=2025-07-19T01:15:01.769+03:00 level=INFO msg="email scheduler started" workers=5 interval=2s visibility_timeout=30s jobs=20 storage=couchbase bucket=scheduler
time=2025-07-19T01:15:01.769+03:00 level=INFO msg="press Ctrl+C to stop gracefully"

# Email jobs processed with Couchbase persistence and type safety
time=2025-07-19T01:15:03.781+03:00 level=INFO msg="sending email" job-id=8d8e35dc-1d89-42d5-86a7-2396fd9e5ea3 to=bob@company.org type=reminder subject="reminder email #10" duration=3.00s
time=2025-07-19T01:15:03.785+03:00 level=INFO msg="sending email" job-id=1616e0f2-1135-4793-bea5-469c18f4584a to=alice@example.com type=reminder subject="reminder email #14" duration=5.00s
time=2025-07-19T01:15:06.781+03:00 level=INFO msg="email sent successfully" job-id=8d8e35dc-1d89-42d5-86a7-2396fd9e5ea3 to=bob@company.org type=reminder
time=2025-07-19T01:15:06.781+03:00 level=INFO msg="job completed" job-id=8d8e35dc-1d89-42d5-86a7-2396fd9e5ea3 worker-id=0 duration=3.00s
time=2025-07-19T01:15:07.791+03:00 level=INFO msg="sending email" job-id=fc7b1ec2-f242-4430-8702-64e06e574d77 to=charlie@startup.io type=notification subject="notification email #3" duration=4.00s
time=2025-07-19T01:15:08.785+03:00 level=INFO msg="email sent successfully" job-id=1616e0f2-1135-4793-bea5-469c18f4584a to=alice@example.com type=reminder
time=2025-07-19T01:15:08.785+03:00 level=INFO msg="job completed" job-id=1616e0f2-1135-4793-bea5-469c18f4584a worker-id=2 duration=5.00s

# Graceful shutdown with Couchbase persistence
time=2025-07-19T01:15:16.039+03:00 level=INFO msg="received shutdown signal" signal=interrupt
time=2025-07-19T01:15:17.837+03:00 level=INFO msg="shutting down scheduler... making remaining jobs visible" remaining-jobs=0
time=2025-07-19T01:15:19.837+03:00 level=INFO msg="scheduler shutdown complete"
time=2025-07-19T01:15:19.837+03:00 level=INFO msg="scheduler stopped gracefully"
```

## Monitoring Jobs

You can inspect email jobs using the Couchbase Web Console or N1QL queries:

### **Web Console**
1. Access http://localhost:8091
2. Go to **Query** tab
3. Run N1QL queries on the `scheduler` bucket

### **N1QL Queries**
```sql
-- View all email jobs
SELECT * FROM scheduler;

-- View pending email jobs
SELECT * FROM scheduler WHERE status = "pending";

-- View completed email jobs  
SELECT * FROM scheduler WHERE status = "completed";

-- View jobs with visibility timeout
SELECT * FROM scheduler WHERE visibleAfter IS NOT NULL;

-- View jobs by email type
SELECT * FROM scheduler WHERE payload.type = "welcome";

-- Count jobs by status
SELECT status, COUNT(*) as count 
FROM scheduler 
GROUP BY status;

-- Count jobs by email type
SELECT payload.type, COUNT(*) as count 
FROM scheduler 
GROUP BY payload.type;
```

## Performance Optimization

### **Recommended Indexes**
Create indexes for optimal query performance:

```sql
-- Primary index (if not exists)
CREATE PRIMARY INDEX ON scheduler;

-- Composite index for job queries
CREATE INDEX idx_job_queries ON scheduler(status, processAfter, visibleAfter);

-- Email type index
CREATE INDEX idx_email_type ON scheduler(payload.type);
```

### **Configuration Tuning**
| Setting | Value | Description |
|---------|-------|-------------|
| **Workers** | 5 | Concurrent goroutines |
| **Fetch Interval** | 2s | Query frequency when idle |
| **Visibility Timeout** | 30s | Fault tolerance window |
| **Jobs** | 20 | Email jobs created |
| **Bucket** | `scheduler` | Couchbase bucket name |
| **Scope** | `jobs` | Custom scope for email jobs |
| **Collection** | `email-jobs` | Collection for email job documents |
| **Memory Quota** | 256MB+ | Minimum bucket memory |

### **Alternative Store Constructors**
Depending on your Couchbase version and requirements:

```go
// Current example: Custom scope with email-jobs collection
store := couchbasestore.NewCouchbaseStore[EmailJob](bucket, "jobs", "email-jobs")

// Alternative: Default scope with custom collection
store := couchbasestore.NewCouchbaseStoreWithCollection[EmailJob](bucket, "jobs")

// Alternative: Different scope/collection for other job types
store := couchbasestore.NewCouchbaseStore[EmailJob](bucket, "production", "background_tasks")
```

## Comparison with Other Stores

| Feature | Memory Store | MongoDB | Couchbase |
|---------|-------------|---------|-----------|
| **Persistence** | ❌ Lost on restart | ✅ Survives restarts | ✅ Survives restarts |
| **Scalability** | ❌ Single instance | ✅ Replica sets | ✅ Auto-scaling clusters |
| **Query Language** | ❌ Go code only | ✅ MongoDB Query | ✅ N1QL (SQL for JSON) |
| **Caching** | ✅ In-memory | ⚠️ Manual caching | ✅ Built-in caching |
| **Setup Complexity** | ✅ Zero setup | ⚠️ Requires MongoDB | ⚠️ Requires Couchbase |
| **Performance** | ✅ Fastest | ✅ Fast with indexes | ✅ Sub-millisecond |
| **Production Ready** | ❌ Development only | ✅ Production ready | ✅ Enterprise ready |

## Troubleshooting

### **Connection Issues**
```bash
# Test Couchbase connection
curl http://localhost:8091/

# Check if Couchbase is running
docker ps | grep couchbase  # Docker
```

### **Bucket Issues**
- Ensure `scheduler` bucket exists
- Check bucket memory quota (minimum 256MB)
- Verify read/write permissions

### **Performance Issues**
- Create appropriate indexes for N1QL queries
- Monitor bucket memory usage
- Adjust worker count based on cluster capacity

### **Authentication Issues**
- Verify username/password combination
- Check RBAC permissions for bucket access
- For Cloud: ensure IP whitelist includes your address

## Next Steps

- **Production Deployment**: Use Couchbase Cloud or multi-node clusters
- **Monitoring**: Add Couchbase metrics and alerting
- **Scaling**: Leverage Couchbase's auto-scaling capabilities
- **Advanced Features**: Explore Couchbase Analytics and Full-Text Search
- **Custom Jobs**: Add typed payloads for specific job types

This example demonstrates production-ready job scheduling with Couchbase's enterprise NoSQL capabilities! 🚀 