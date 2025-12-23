## 🚀 Load Test Results (Production Grade Proof)

Tested using **k6** with **1,000 Concurrent Users** simulating a Flash Sale war.

### 📊 Summary

- **Total Requests:** ~179,000 requests in 30 seconds.
- **Throughput:** ~6,000 requests/second.
- **Latency (Avg):** 1.62ms ⚡
- **Success Rate:** 100% (No 500 Internal Server Errors).

### 🛡️ Data Consistency Check

| Metric                                  | Count     | Status     |
| :-------------------------------------- | :-------- | :--------- |
| Initial Stock                           | 5,000     | -          |
| **Successful API Responses (HTTP 202)** | **5,000** | ✅ Match   |
| **Data Persisted in PostgreSQL**        | **5,000** | ✅ Match   |
| Overselling Count                       | 0         | ✅ Perfect |

_Note: The high number of HTTP 409 responses in k6 results indicates successful "Sold Out" handling logic for 174k+ late requests._
