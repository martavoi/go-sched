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

- **Creates 50 jobs** with random scheduling (1-30 seconds delay)
- **5 concurrent workers** process jobs with random durations (1-8 seconds)
- **30-second visibility timeout** for automatic fault recovery
- **MongoDB persistence** - jobs survive application restarts
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

The example creates documents in the `scheduler.jobs` collection:

```javascript
{
  "_id": "job-1",
  "status": "pending",           // "pending" or "completed"
  "processAfter": ISODate("..."), // When job should run
  "visibleAfter": ISODate("..."), // Visibility timeout (null = visible)
  "processedAt": ISODate("..."),  // When job completed (null = not completed)
  "payload": null                 // Job data (any type)
}
```

## Sample Execution Log

```
time=2025-07-18T15:30:45.123+03:00 level=INFO msg="connected to MongoDB" uri="mongodb://localhost:27017"
time=2025-07-18T15:30:45.234+03:00 level=INFO msg="scheduler started" workers=5 interval=2s visibility_timeout=30s jobs=50 storage=mongodb
time=2025-07-18T15:30:45.234+03:00 level=INFO msg="press Ctrl+C to stop gracefully"

# Jobs processed with MongoDB persistence
time=2025-07-18T15:30:47.123+03:00 level=INFO msg="processing job" job-id=job-15 duration=3s
time=2025-07-18T15:30:47.124+03:00 level=INFO msg="processing job" job-id=job-23 duration=5s
time=2025-07-18T15:30:50.123+03:00 level=INFO msg="job completed" job-id=job-15
time=2025-07-18T15:30:52.124+03:00 level=INFO msg="job completed" job-id=job-23

# Graceful shutdown with MongoDB
time=2025-07-18T15:31:15.456+03:00 level=INFO msg="received shutdown signal" signal=interrupt
time=2025-07-18T15:31:15.500+03:00 level=INFO msg="shutting down scheduler... making remaining jobs visible" remaining-jobs=3
time=2025-07-18T15:31:18.500+03:00 level=INFO msg="scheduler shutdown complete"
time=2025-07-18T15:31:18.500+03:00 level=INFO msg="scheduler stopped gracefully"
```

## Monitoring Jobs

You can inspect jobs directly in MongoDB:

```javascript
// Connect to MongoDB
use scheduler

// View all jobs
db.jobs.find().pretty()

// View pending jobs
db.jobs.find({status: "pending"}).pretty()

// View completed jobs
db.jobs.find({status: "completed"}).pretty()

// View jobs with visibility timeout
db.jobs.find({visibleAfter: {$exists: true}}).pretty()

// Count jobs by status
db.jobs.aggregate([
  {$group: {_id: "$status", count: {$sum: 1}}}
])
```

## Performance Optimization

### **Recommended Indexes**
```javascript
// Create indexes for optimal query performance
db.jobs.createIndex({status: 1, processAfter: 1, visibleAfter: 1})
db.jobs.createIndex({status: 1, visibleAfter: 1})
```

### **Configuration Tuning**
| Setting | Value | Description |
|---------|-------|-------------|
| **Workers** | 5 | Concurrent goroutines |
| **Fetch Interval** | 2s | Query frequency when idle |
| **Visibility Timeout** | 30s | Fault tolerance window |
| **Database** | `scheduler` | MongoDB database name |
| **Collection** | `jobs` | MongoDB collection name |

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