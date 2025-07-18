# MongoDB Job Scheduler Example

This example demonstrates the go-sched library with **MongoDB persistence**, showcasing real database storage with the same concurrent job processing and fault tolerance features.

## Prerequisites

### 1. **MongoDB Server**
You need a running MongoDB instance. Choose one:

**Option A: Local MongoDB**
```bash
# Install MongoDB locally and start it
mongod --dbpath ./data/db
```

**Option B: Docker**
```bash
# Run MongoDB in Docker
docker run -d -p 27017:27017 --name mongo-scheduler mongo:latest
```

**Option C: MongoDB Atlas**
- Create free cluster at [mongodb.com/atlas](https://mongodb.com/atlas)
- Get connection string

### 2. **Go Dependencies**
```bash
# Install MongoDB driver (if not already installed)
go mod tidy
```

## Configuration

### **Environment Variables**
```bash
# Optional: Set MongoDB connection string
export MONGO_URI="mongodb://localhost:27017"

# For MongoDB Atlas:
export MONGO_URI="mongodb+srv://username:password@cluster.mongodb.net/"

# For Docker with auth:
export MONGO_URI="mongodb://username:password@localhost:27017"
```

**Default**: If `MONGO_URI` is not set, uses `mongodb://localhost:27017`

## Running the Example

```bash
cd examples/mongo
go run main.go
```

## What This Example Does

- **Creates 20 email jobs** with random scheduling (1-30 seconds delay)
- **5 concurrent workers** process emails with random durations (1-5 seconds)
- **30-second visibility timeout** for automatic fault recovery
- **MongoDB persistence** - jobs survive application restarts
- **Type-safe email processing** - demonstrates structured payloads
- **Demonstrates graceful shutdown** with immediate job visibility reset

## Key Features Demonstrated

### 1. **Persistent Storage**
- Jobs are stored in MongoDB collection `scheduler.jobs`
- Survives application restarts and crashes
- Visibility timeout ensures fault tolerance across restarts

### 2. **MongoDB Integration**
- Uses MongoDB driver with proper connection handling
- Efficient queries with indexes on key fields
- Error handling for database operations

### 3. **Production Patterns**
- Environment-based configuration
- Connection validation and error handling
- Graceful database disconnection

### 4. **Fault Tolerance**
- Jobs automatically become visible again after 30s if worker crashes
- Database persistence ensures no job loss on application restart
- Graceful shutdown makes remaining jobs immediately available

## Database Schema

The example creates documents in the `scheduler.email_jobs` collection:

### **Pending Job Document**
```javascript
{
  "_id": "4a9188cf-0d3e-45fe-9e19-1aa5e13c2745",
  "status": "pending",
  "processAfter": {
    "$date": "2025-07-18T21:53:23.722Z"
  },
  "payload": {
    "to": "bob@company.org",
    "subject": "notification email #1",
    "type": "notification"
  }
}
```

### **Completed Job Document**
```javascript
{
  "_id": "a5c64868-8a15-4449-848b-2df0a71c3280",
  "status": "completed",
  "processAfter": {
    "$date": "2025-07-18T21:53:13.733Z"
  },
  "payload": {
    "to": "alice@example.com",
    "subject": "reminder email #2",
    "type": "reminder"
  },
  "processedAt": {
    "$date": "2025-07-18T21:53:16.821Z"
  },
  "visibleAfter": null
}
```

### **Field Descriptions**
| Field | Type | Description |
|-------|------|-------------|
| `_id` | String | Unique job identifier (UUID) |
| `status` | String | Job state: `"pending"` or `"completed"` |
| `processAfter` | Date | When job should be processed |
| `visibleAfter` | Date/null | Visibility timeout (null = visible) |
| `processedAt` | Date/null | When job completed (only for completed jobs) |
| `payload` | Object | EmailJob data with `to`, `subject`, `type` |

## Sample Execution Log

```
time=2025-07-19T00:53:01.721+03:00 level=INFO msg="connected to MongoDB" uri=mongodb://subly:P%40ssw0rd!@localhost:27017
time=2025-07-19T00:53:01.769+03:00 level=INFO msg="email scheduler started" workers=5 interval=2s visibility_timeout=30s jobs=20 storage=mongodb collection=email_jobs
time=2025-07-19T00:53:01.769+03:00 level=INFO msg="press Ctrl+C to stop gracefully"

# Email jobs processed with MongoDB persistence and type safety
time=2025-07-19T00:53:03.781+03:00 level=INFO msg="sending email" job-id=8d8e35dc-1d89-42d5-86a7-2396fd9e5ea3 to=bob@company.org type=reminder subject="reminder email #10" duration=3.00s
time=2025-07-19T00:53:03.785+03:00 level=INFO msg="sending email" job-id=1616e0f2-1135-4793-bea5-469c18f4584a to=alice@example.com type=reminder subject="reminder email #14" duration=5.00s
time=2025-07-19T00:53:06.781+03:00 level=INFO msg="email sent successfully" job-id=8d8e35dc-1d89-42d5-86a7-2396fd9e5ea3 to=bob@company.org type=reminder
time=2025-07-19T00:53:06.781+03:00 level=INFO msg="job completed" job-id=8d8e35dc-1d89-42d5-86a7-2396fd9e5ea3 worker-id=0 duration=3.00s
time=2025-07-19T00:53:07.791+03:00 level=INFO msg="sending email" job-id=fc7b1ec2-f242-4430-8702-64e06e574d77 to=charlie@startup.io type=notification subject="notification email #3" duration=4.00s
time=2025-07-19T00:53:08.785+03:00 level=INFO msg="email sent successfully" job-id=1616e0f2-1135-4793-bea5-469c18f4584a to=alice@example.com type=reminder
time=2025-07-19T00:53:08.785+03:00 level=INFO msg="job completed" job-id=1616e0f2-1135-4793-bea5-469c18f4584a worker-id=2 duration=5.00s

# Graceful shutdown with MongoDB persistence
time=2025-07-19T00:53:16.039+03:00 level=INFO msg="received shutdown signal" signal=interrupt
time=2025-07-19T00:53:17.837+03:00 level=INFO msg="shutting down scheduler... making remaining jobs visible" remaining-jobs=0
time=2025-07-19T00:53:19.837+03:00 level=INFO msg="scheduler shutdown complete"
time=2025-07-19T00:53:19.837+03:00 level=INFO msg="scheduler stopped gracefully"
```

## Monitoring Jobs

You can inspect email jobs directly in MongoDB:

```javascript
// Connect to MongoDB
use scheduler

// View all email jobs
db.email_jobs.find().pretty()

// View pending email jobs
db.email_jobs.find({status: "pending"}).pretty()

// View completed email jobs
db.email_jobs.find({status: "completed"}).pretty()

// View jobs with visibility timeout
db.email_jobs.find({visibleAfter: {$exists: true}}).pretty()

// View jobs by email type
db.email_jobs.find({"payload.type": "welcome"}).pretty()

// Count jobs by status
db.email_jobs.aggregate([
  {$group: {_id: "$status", count: {$sum: 1}}}
])

// Count jobs by email type
db.email_jobs.aggregate([
  {$group: {_id: "$payload.type", count: {$sum: 1}}}
])
```

## Performance Optimization

### **Recommended Indexes**
```javascript
// Create indexes for optimal query performance
db.email_jobs.createIndex({status: 1, processAfter: 1, visibleAfter: 1})
db.email_jobs.createIndex({status: 1, visibleAfter: 1})
db.email_jobs.createIndex({"payload.type": 1})
```

### **Configuration Tuning**
| Setting | Value | Description |
|---------|-------|-------------|
| **Workers** | 5 | Concurrent goroutines |
| **Fetch Interval** | 2s | Query frequency when idle |
| **Visibility Timeout** | 30s | Fault tolerance window |
| **Jobs** | 20 | Email jobs created |
| **Database** | `scheduler` | MongoDB database name |
| **Collection** | `email_jobs` | MongoDB collection name |

## Comparison with Memory Store

| Feature | Memory Store | MongoDB Store |
|---------|-------------|---------------|
| **Persistence** | ❌ Lost on restart | ✅ Survives restarts |
| **Scalability** | ❌ Single instance | ✅ Multiple instances |
| **Query Performance** | ✅ Instant | ✅ Fast with indexes |
| **Setup Complexity** | ✅ Zero setup | ⚠️ Requires MongoDB |
| **Production Ready** | ❌ Development only | ✅ Production ready |

## Troubleshooting

### **Connection Issues**
```bash
# Test MongoDB connection
mongo mongodb://localhost:27017

# Check if MongoDB is running
ps aux | grep mongod  # Linux/Mac
Get-Process -Name mongod  # Windows PowerShell
```

### **Permission Issues**
```bash
# MongoDB Atlas: Check IP whitelist and credentials
# Local MongoDB: Ensure no authentication required or configure auth
```

### **Performance Issues**
```bash
# Create indexes for better query performance
# Monitor MongoDB logs for slow queries
# Adjust worker count based on database capacity
```

## Next Steps

- **Production Deployment**: Use MongoDB Atlas or replica sets
- **Monitoring**: Add MongoDB metrics and alerting
- **Scaling**: Run multiple scheduler instances
- **Custom Jobs**: Add typed payloads for specific job types

This example demonstrates production-ready job scheduling with MongoDB persistence! 🚀 