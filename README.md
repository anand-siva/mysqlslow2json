# mysqlslow2json

`mysqlslow2json` is a small Go CLI that converts MySQL slow query logs into newline-delimited JSON (`.jsonl`).

It parses each slow query entry and emits a JSON object with:

- `time`
- `query_time`
- `lock_time`
- `rows_sent`
- `rows_examined`
- `database`
- `set_timestamp`
- `sql`

This is useful if you want to:

- inspect slow query logs more easily
- feed them into another tool or script
- ship them into a log pipeline later

## Why It Is Useful

Raw MySQL slow query logs are readable, but they are awkward to analyze programmatically. Each entry spans multiple lines, optional fields appear inconsistently, and it is not convenient to filter or transform with standard tooling.

`mysqlslow2json` turns each slow query into a single JSON object so you can:

- search for long-running queries by `query_time`
- sort or filter by `rows_examined`
- pipe the output into `jq`
- load the data into another service or analytics workflow
- keep the original SQL text attached to the timing metadata

## Features

- parses standard MySQL slow query log blocks
- skips the startup header at the top of the log
- writes one JSON object per line
- supports a configurable output file
- uses a sensible default output path

## Usage

Run the converter with a required slow query log path:

```bash
go run ./cmd/mysqlslow2json --slow-query-log sample_slow_query.log
```

By default, output is written to:

```bash
slow-query.jsonl
```

You can also override the output file:

```bash
go run ./cmd/mysqlslow2json \
  --slow-query-log sample_slow_query.log \
  --output-file out.jsonl
```

## Docker Setup

This repo includes a local MySQL setup in [compose.yml](/Users/amoney/mysqlslow2json/compose.yml), a MySQL config in [my.cnf](/Users/amoney/mysqlslow2json/my.cnf), and seed data in [001-seed.sql](/Users/amoney/mysqlslow2json/mysql_init/001-seed.sql).

Start MySQL with Docker Compose:

```bash
docker compose up -d
```

Check that the container is running:

```bash
docker compose ps
```

The MySQL config enables the slow query log and writes it to:

```bash
logs/mysql-slow.log
```

On first startup, Docker will load the seed SQL from `mysql_init/001-seed.sql` and create:

- database: `sample_app`
- user: `appuser`
- password: `apppass`
- sample `users` and `orders` tables

If you have already started the container before fixing the seed mount, remove the existing volume and start fresh:

```bash
docker compose down -v
docker compose up -d
```

## Flags

- `--slow-query-log`: path to the MySQL slow query log file
- `--output-file`: path to the output JSONL file, default is `slow-query.jsonl`
- `--follow`: continue reading the slow query log as it grows

## Example

Sample input:

```log
# Time: 2026-03-06T08:10:00.123456Z
# User@Host: app_user[app_user] @  [192.168.1.10]  Id:    101
# Query_time: 2.500200  Lock_time: 0.000050 Rows_sent: 50  Rows_examined: 1250000
use ecom_db;
SET timestamp=1772784600;
SELECT c.name, o.order_date, o.total FROM customers c JOIN orders o ON c.id = o.customer_id WHERE o.status = 'PENDING' ORDER BY o.order_date DESC;
```

Command:

```bash
go run ./cmd/mysqlslow2json --slow-query-log sample_slow_query.log
```

Output written to `slow-query.jsonl`:

```json
{"time":"2026-03-06T08:10:00.123456Z","query_time":2.5002,"lock_time":0.00005,"rows_sent":50,"rows_examined":1250000,"database":"ecom_db","set_timestamp":1772784600,"sql":"SELECT c.name, o.order_date, o.total FROM customers c JOIN orders o ON c.id = o.customer_id WHERE o.status = 'PENDING' ORDER BY o.order_date DESC; "}
{"time":"2026-03-06T08:15:22.987654Z","query_time":15.23411,"lock_time":0.0123,"rows_sent":0,"rows_examined":500000,"database":"","set_timestamp":1772784922,"sql":"UPDATE products SET stock = stock - 1 WHERE category_id IN (SELECT id FROM categories WHERE name LIKE '%Electronics%') AND status = 'active'; "}
{"time":"2026-03-06T08:22:11.111222Z","query_time":8.45,"lock_time":0.005,"rows_sent":0,"rows_examined":850000,"database":"ecom_db","set_timestamp":1772785331,"sql":"DELETE FROM session_logs WHERE last_accessed \u003c DATE_SUB(NOW(), INTERVAL 30 DAY); "}
```

## Before And After

Before, a slow query entry looks like this:

```log
# Time: 2026-03-06T08:35:45.333444Z
# User@Host: reporting[reporting] @  [192.168.1.50]  Id:    302
# Query_time: 45.102300  Lock_time: 0.000100 Rows_sent: 1  Rows_examined: 9500000
SET timestamp=1772786145;
SELECT COUNT(DISTINCT user_id), SUM(amount) FROM transactions WHERE created_at BETWEEN '2025-01-01' AND '2025-12-31' AND status = 'SUCCESS';
```

After, the same entry becomes one JSON line:

```json
{"time":"2026-03-06T08:35:45.333444Z","query_time":45.1023,"lock_time":0.0001,"rows_sent":1,"rows_examined":9500000,"database":"","set_timestamp":1772786145,"sql":"SELECT COUNT(DISTINCT user_id), SUM(amount) FROM transactions WHERE created_at BETWEEN '2025-01-01' AND '2025-12-31' AND status = 'SUCCESS'; "}
```

That makes simple analysis much easier. For example, you can quickly inspect the heaviest queries:

```bash
cat slow-query.jsonl | jq 'select(.query_time > 10)'
```

## Output Format

Each line in the output file is a standalone JSON object. That makes the output easy to:

- stream
- grep
- process with `jq`
- ingest into log tools later

Example:

```bash
cat slow-query.jsonl | jq .
```

## Testing Follow Mode

You can test follow mode against a local MySQL slow query log and watch new JSON entries appear in real time.

Before starting follow mode, bring up MySQL:

```bash
docker compose up -d
```

Start the converter:

```bash
go run ./cmd/mysqlslow2json --slow-query-log logs/mysql-slow.log --follow
```

In another terminal, connect to MySQL:

```bash
mysql -h 127.0.0.1 -P 3306 -uappuser -papppass sample_app
```

Run a deliberately slow query:

```sql
select sleep(2);
```

You can also run slow queries against the seeded sample data:

```sql
SELECT u.id, u.first_name, o.order_total, SLEEP(2)
FROM users u
JOIN orders o ON o.user_id = u.id;
```

```sql
SELECT COUNT(*), SLEEP(2)
FROM orders
WHERE status = 'paid';
```

```sql
SELECT u.email, SUM(o.order_total), SLEEP(2)
FROM users u
JOIN orders o ON o.user_id = u.id
WHERE o.status IN ('paid', 'pending')
GROUP BY u.email;
```

Expected MySQL output:

```text
+----------+
| sleep(2) |
+----------+
|        0 |
+----------+
1 row in set (2.003 sec)
```

In a third terminal, watch the generated JSON lines file:

```bash
tail -f slow-query.jsonl
```

Once MySQL writes the slow query log entry, you should see a new JSON object appear in `slow-query.jsonl`.

```
{"time":"2026-04-01T23:21:11.833525Z","query_time":8.018268,"lock_time":0.000007,"rows_sent":4,"rows_examined":7,"database":"sample_app","set_timestamp":1775085663,"sql":"SELECT u.id, u.first_name, o.order_total, SLEEP(2) FROM users u JOIN orders o ON o.user_id = u.id; ","user":"appuser","host":"192.168.65.1","thread_id":8}
{"time":"2026-04-01T23:21:38.889907Z","query_time":8.006375,"lock_time":0.000005,"rows_sent":4,"rows_examined":7,"database":"","set_timestamp":1775085690,"sql":"SELECT u.id, u.first_name, o.order_total, SLEEP(2) FROM users u JOIN orders o ON o.user_id = u.id; ","user":"appuser","host":"192.168.65.1","thread_id":8}
```

## Notes

- entries without a `use ...;` line will have an empty `database`
- SQL is currently stored as a single string
- in `--follow` mode, the current block stays buffered until the next `# Time:` line arrives
- the output file is recreated on startup, not appended to

## Coming Next

- support for parsing additional slow log metadata such as user, host, and connection id
- optional pretty-printed JSON output for debugging
- filtering by query time or rows examined
- direct export to downstream systems or analysis pipelines
