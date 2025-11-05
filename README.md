# Telemetry Infrastructure

## Performance Results

### Setup

To run the performance testing tool, you need to install `hey`:

```bash
go install github.com/rakyll/hey@latest
```

### Running Performance Tests

To run a performance test:

```bash
hey -n 5000 -c 50 http://localhost:8080/telemetry/<EP>
```

### Results

Results for 5000 requests with 50 client concurrency using the command:
```bash
hey -n 5000 -c 50 http://localhost:8080/telemetry/ListMetrics
```

```
Summary:
  Total:	1.5811 secs
  Slowest:	0.0586 secs
  Fastest:	0.0024 secs
  Average:	0.0156 secs
  Requests/sec:	3162.3278

Response time histogram:
  0.002 [1]	|
  0.008 [15]	|
  0.014 [1392]	|■■■■■■■■■■■■■■■■■■
  0.019 [3127]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.025 [383]	|■■■■■
  0.031 [37]	|
  0.036 [1]	|
  0.042 [7]	|
  0.047 [20]	|
  0.053 [14]	|
  0.059 [3]	|

Latency distribution:
  10% in 0.0121 secs
  25% in 0.0134 secs
  50% in 0.0150 secs
  75% in 0.0170 secs
  90% in 0.0191 secs
  95% in 0.0209 secs
  99% in 0.0297 secs

Details (average, fastest, slowest):
  DNS+dialup:	0.0001 secs, 0.0024 secs, 0.0586 secs
  DNS-lookup:	0.0000 secs, 0.0000 secs, 0.0031 secs
  req write:	0.0000 secs, 0.0000 secs, 0.0042 secs
  resp wait:	0.0154 secs, 0.0024 secs, 0.0478 secs
  resp read:	0.0000 secs, 0.0000 secs, 0.0010 secs

Status code distribution:
  [200]	5000 responses
```

**Performance Optimizations Applied:**
- Replaced blocking Redis `KEYS` command with non-blocking `SCAN`
- Batch fetching using Redis Pipeline (single round-trip instead of N queries)
- Streaming JSON encoding to reduce memory allocation

---

Results for 5000 requests with 50 client concurrency using the command:
```bash
hey -n 5000 -c 50 "http://localhost:8080/telemetry/GetMetric?switch_id=sw5&metric=latency_ms"
```

```
Summary:
  Total:	0.3281 secs
  Slowest:	0.0130 secs
  Fastest:	0.0008 secs
  Average:	0.0032 secs
  Requests/sec:	15237.1506

  Total data:	30000 bytes
  Size/request:	6 bytes

Response time histogram:
  0.001 [1]	|
  0.002 [311]	|■■■■■
  0.003 [2752]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.004 [1590]	|■■■■■■■■■■■■■■■■■■■■■■■
  0.006 [274]	|■■■■
  0.007 [15]	|
  0.008 [16]	|
  0.009 [16]	|
  0.011 [10]	|
  0.012 [11]	|
  0.013 [4]	|


Latency distribution:
  10% in 0.0022 secs
  25% in 0.0026 secs
  50% in 0.0030 secs
  75% in 0.0037 secs
  90% in 0.0043 secs
  95% in 0.0047 secs
  99% in 0.0077 secs

Details (average, fastest, slowest):
  DNS+dialup:	0.0001 secs, 0.0008 secs, 0.0130 secs
  DNS-lookup:	0.0000 secs, 0.0000 secs, 0.0021 secs
  req write:	0.0000 secs, 0.0000 secs, 0.0021 secs
  resp wait:	0.0030 secs, 0.0008 secs, 0.0075 secs
  resp read:	0.0001 secs, 0.0000 secs, 0.0022 secs

Status code distribution:
  [200]	5000 responses
```
