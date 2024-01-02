### Repository

- https://github.com/finnng/go-pg-bench

### Prerequisites

- Go version 1.21.0
- docker-compose (you may need to update docker-compose.yml for intel based computer)

### Steps

1. Pull the source code
2. Start the databases: `docker compose up -d`
3. Create a Postgres test database. Use any database client to create a database name `test` and grant permission for
   the default user `postgres` on it.

```sql
CREATE DATABASE test;
GRANT ALL PRIVILEGES ON DATABASE test TO postgres;
```

1. Start the API server, it should automatically provision the tables. From the repo’s root directory, type

```bash
go run api-server/app.go
```

1. Start other workers to complete the full system, open other terminal tabs for these commands

```bash
go run worker-due-job-checker/app.go
```

```bash
go run worker-job-fixer/app.go
```

```bash
go run data-feed/app.go
```

### Monitoring

I haven’t handled the Grafana database migration yet, so you need to head to the Grafana dashboard
at [`http://localhost:3000`](http://localhost:3000/) according to the docker-compose file Grafana port.

1. Setup Prometheus as the data source
2. Play around with the metrics sent from the scheduling system
