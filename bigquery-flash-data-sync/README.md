# BigQuery Flash Data Sync

A Go application that synchronizes SQL databases (MySQL, PostgreSQL) to Google Cloud BigQuery with automatic schema inference, concurrent processing, and dynamic multi-database configuration.

## 📋 Quick Start

### Prerequisites

- Go 1.21+
- Google Cloud SDK with BigQuery API enabled
- MySQL 5.7+ or PostgreSQL 12+ with read access
- Service account with `bigquery.dataEditor` and `bigquery.jobUser` roles

### Installation

```bash
# Clone and navigate
cd bigquery-flash-data-sync

# Install dependencies
go mod download

# Set up authentication
gcloud auth application-default login
gcloud config set project YOUR_PROJECT_ID

# Configure environment
cp .env.example .env
# Edit .env with your credentials

# Build
go build -o bin/datasync cmd/datasync/main.go
```

### Run

```bash
# Development
go run ./cmd/datasync

# Production (using built binary)
./bin/datasync
```

## ⚙️ Configuration

Copy `.env.example` to `.env` and configure your databases dynamically.

### Minimal Configuration

```bash
# Google Cloud / BigQuery
GCP_PROJECT_ID=my-gcp-project-123
BQ_DATASET_ID=analytics_data

# Databases to sync (comma-separated identifiers)
SYNC_DATABASES=finance,salesforce

# Finance Database
FINANCE_ENABLED=true
FINANCE_DB_TYPE=mysql
FINANCE_DB_HOST=finance-db.example.com
FINANCE_DB_PORT=3306
FINANCE_DB_NAME=finance_prod
FINANCE_DB_USER=reader
FINANCE_DB_PASSWORD=secret
FINANCE_TABLES=invoices,payments,accounts

# Salesforce Database
SALESFORCE_ENABLED=true
SALESFORCE_DB_TYPE=mysql
SALESFORCE_DB_HOST=salesforce-db.example.com
SALESFORCE_DB_PORT=3306
SALESFORCE_DB_NAME=salesforce_mirror
SALESFORCE_DB_USER=reader
SALESFORCE_DB_PASSWORD=secret
SALESFORCE_TABLES=opportunities,contacts
```

### Configuration Reference

| Variable                  | Description                                   | Default      |
| ------------------------- | --------------------------------------------- | ------------ |
| `GCP_PROJECT_ID`          | Google Cloud Project ID                       | _required_   |
| `BQ_DATASET_ID`           | BigQuery dataset ID                           | _required_   |
| `SYNC_DATABASES`          | Comma-separated database identifiers          | _required_   |
| `DB_TYPE`                 | Default database type (`mysql` or `postgres`) | `mysql`      |
| `DB_HOST`                 | Default database host                         | `localhost`  |
| `DB_PORT`                 | Default database port                         | `3306`       |
| `DB_MAX_OPEN_CONNECTIONS` | Max concurrent connections                    | `15`         |
| `DB_MAX_IDLE_CONNECTIONS` | Max idle connections                          | `5`          |
| `DB_CONN_MAX_LIFETIME`    | Connection lifetime                           | `1m`         |
| `SYNC_TIMEOUT`            | Total sync timeout                            | `10m`        |
| `DATE_FORMAT`             | Timestamp format                              | `2006-01-02` |
| `DEFAULT_BATCH_SIZE`      | Rows per batch                                | `1000`       |
| `DRY_RUN`                 | Test without writing to BigQuery              | `false`      |
| `AUTO_CREATE_TABLES`      | Auto-create BigQuery tables                   | `true`       |
| `TRUNCATE_ON_SYNC`        | Delete data before sync                       | `false`      |
| `LOG_ENV`                 | Logging format (`dev` or `prod`)              | `dev`        |
| `LOG_LEVEL`               | Log level                                     | `info`       |

### Per-Database Configuration

For each database in `SYNC_DATABASES`, use the pattern `{DATABASE_ID}_SETTING`:

```bash
# Example: FINANCE database
FINANCE_ENABLED=true
FINANCE_DB_TYPE=mysql
FINANCE_DB_HOST=db.example.com
FINANCE_DB_PORT=3306
FINANCE_DB_NAME=finance_prod
FINANCE_DB_USER=reader
FINANCE_DB_PASSWORD=secret
FINANCE_TABLES=table1,table2,table3
```

### Per-Table Configuration (Optional)

For fine-grained control, use the pattern `{DATABASE_ID}_{TABLE_NAME}_SETTING`:

```bash
FINANCE_INVOICES_ENABLED=true
FINANCE_INVOICES_TARGET_TABLE=finance_invoices
FINANCE_INVOICES_PRIMARY_KEY=invoice_id
FINANCE_INVOICES_TIMESTAMP_COLUMN=updated_at
FINANCE_INVOICES_COLUMNS=id,amount,status
FINANCE_INVOICES_BATCH_SIZE=5000
```

See [`.env.example`](.env.example) for complete documentation.

## 🏗 How It Works

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  MySQL/Postgres │────▶│  Schema Inference │────▶│    BigQuery     │
│   Databases     │     │  & Data Extract   │     │    Dataset      │
└─────────────────┘     └──────────────────┘     └─────────────────┘
```

1. **Configuration Loading**: Reads environment variables and builds database/table configs
2. **Schema Inference**: Automatically detects source schemas and maps to BigQuery types
3. **Concurrent Processing**: Parallel extraction and loading of multiple tables
4. **Data Sanitization**: Handles special characters, NULLs, and invalid UTF-8
5. **BigQuery Loading**: Creates/updates tables and loads data with `WRITE_TRUNCATE`

### Supported Type Mappings

| MySQL Type             | PostgreSQL Type | BigQuery Type |
| ---------------------- | --------------- | ------------- |
| VARCHAR, TEXT          | VARCHAR, TEXT   | STRING        |
| INT, BIGINT            | INTEGER, BIGINT | INTEGER       |
| FLOAT, DOUBLE, DECIMAL | FLOAT, NUMERIC  | FLOAT         |
| DATE                   | DATE            | DATE          |
| TIME                   | TIME            | TIME          |
| DATETIME, TIMESTAMP    | TIMESTAMP       | TIMESTAMP     |
| BOOLEAN                | BOOLEAN         | BOOLEAN       |
| BLOB, BINARY           | BYTEA           | BYTES         |
| JSON                   | JSON, JSONB     | JSON          |

## 📁 Project Structure

```
bigquery-flash-data-sync/
├── README.md                    # This file
├── .env.example                 # Configuration template
├── go.mod                       # Go module dependencies
├── go.sum                       # Dependency checksums
├── assets/
│   └── schemas/                 # Schema documentation
├── cmd/
│   └── datasync/
│       └── main.go              # Application entry point
└── internal/
    ├── config/
    │   └── config.go            # Environment variable loading
    ├── logger/
    │   └── logger.go            # Structured logging (zap)
    ├── model/
    │   ├── models.go            # Data structures & types
    │   └── parser.go            # Row parsing & type conversion
    └── pipeline/
        ├── bqsetup.go           # Schema inference & table management
        └── job.go               # ETL job orchestration
```

## 🔧 Adding New Databases

Simply update your `.env` file:

```bash
# 1. Add to SYNC_DATABASES
SYNC_DATABASES=finance,salesforce,inventory

# 2. Configure the new database
INVENTORY_ENABLED=true
INVENTORY_DB_TYPE=postgres
INVENTORY_DB_HOST=inventory-db.example.com
INVENTORY_DB_PORT=5432
INVENTORY_DB_NAME=inventory_prod
INVENTORY_DB_USER=reader
INVENTORY_DB_PASSWORD=secret
INVENTORY_TABLES=products,stock_levels,warehouses
```

No code changes required! Schema detection is automatic.

## 🐛 Troubleshooting

### Connection Issues

```bash
# Verify MySQL connectivity
mysql -h $DB_HOST -P $DB_PORT -u $USER -p -e "SHOW TABLES"

# Verify PostgreSQL connectivity
psql -h $DB_HOST -p $DB_PORT -U $USER -d $DB_NAME -c "\dt"

# Check BigQuery access
bq ls --project_id=$GCP_PROJECT_ID $BQ_DATASET_ID
```

### Enable Debug Logging

```bash
LOG_LEVEL=debug go run ./cmd/datasync
```

### Common Errors

| Error                                  | Solution                                          |
| -------------------------------------- | ------------------------------------------------- |
| `Table 'database.table' doesn't exist` | Check table names in `{DB}_TABLES` variable       |
| `dial tcp: i/o timeout`                | Verify `DB_HOST` and `DB_PORT`, check firewall    |
| `Access denied`                        | Verify credentials and user permissions           |
| `Permission denied` (BigQuery)         | Add `bigquery.dataEditor` role to service account |
| `invalid character`                    | Enable debug mode, check for invalid UTF-8 data   |
| `context deadline exceeded`            | Increase `SYNC_TIMEOUT` value                     |

### Test Mode

Run without writing to BigQuery:

```bash
DRY_RUN=true go run ./cmd/datasync
```

## 📊 Performance

| Rows | Columns | Tables | Sync Time | Memory |
| ---- | ------- | ------ | --------- | ------ |
| 1K   | 10      | 5      | ~3s       | ~50MB  |
| 50K  | 25      | 10     | ~20s      | ~200MB |
| 500K | 50      | 15     | ~120s     | ~800MB |

### Optimization Tips

- Increase `DB_MAX_OPEN_CONNECTIONS` for more parallelism
- Adjust `DEFAULT_BATCH_SIZE` based on row size
- Set appropriate `SYNC_TIMEOUT` for large datasets
- Use `{TABLE}_COLUMNS` to sync only needed columns

## 🔒 Security Best Practices

- Never commit `.env` to version control (add to `.gitignore`)
- Use read-only database users with minimal permissions
- Store production credentials in a secret manager (e.g., Google Secret Manager)
- Enable TLS/SSL for all database connections
- Rotate credentials regularly
- Use service accounts with least-privilege IAM roles

## 📝 License

Copyright (c) 2025, WSO2 LLC. All Rights Reserved.

This software is the property of WSO2 LLC. and its suppliers, if any.
Dissemination of any information or reproduction of any material contained
herein in any form is strictly forbidden, unless permitted by WSO2 expressly.
You may not alter or remove any copyright or other notice from copies of this content.

---

**Maintained by**: WSO2 Internal Apps Team
