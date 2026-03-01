# dbcompare - Database Comparison and Migration Tool

A comprehensive command-line tool for comparing and migrating data between different database systems, primarily designed for Oracle↔MySQL, Oracle↔TiDB and MySQL↔MySQL database comparisons.

## Overview

dbcompare is a versatile database comparison tool that enables:

- **SQL Query Result Comparison**: Execute identical SQL queries on different database systems and compare results to verify compatibility
- **Database Migration Assistance**: Facilitate data migration processes between Oracle, MySQL, and TiDB 
- **Schema and Data Consistency Checks**: Validate data integrity after database migrations
- **Automated Data Transfer**: Streamline bulk data movement with configurable processes

## Features

### 1. SQL Comparison (sql-diff)
- Execute the same SQL query on source and destination databases
- Compare query results to validate equivalence 
- Support for multiple database types (MySQL, Oracle, TiDB)
- Input methods: command line, file, or configuration-based SQL statements

### 2. Data Synchronization Control (sync-diff)
- Parallel synchronization of data between databases with configurable concurrency
- Automated configuration file generation for TiDB Lightining
- Progress tracking and status monitoring
- Support for snapshot-based consistent reads
- Multi-threaded processing with configurable concurrency

### 3. Oracle to Target Initialization (o2t-init)
- Bulk export of Oracle data to CSV files using sqluldr2
- Configuration file generation for TiDB Lightning
- Loading data into target databases with progress monitoring
- Row count validation capabilities  
- Multi-stage processing: dump → generate config → load → validate

### 4. Target to Oracle Initialization (t2o-init)
- Dump data from MySQL/TiDB using Dumpling 
- Generate Oracle SQL*Loader control files automatically
- Load data into Oracle using SQL*Loader with parallel processing
- Configurable truncation and pre-processing options
- Comprehensive error handling and logging

## Supported Databases

- **MySQL** (and compatible systems like TiDB via MySQL protocol)
- **Oracle** (using godror driver)
- **TiDB** (as MySQL-compatible target)

## Architecture

### Core Components
- **CLI Layer**: Built with Cobra framework for intuitive command-line interface
- **Database Abstraction**: Unified interface supporting multiple database types
- **Configuration Management**: Flexible YAML and TOML configuration support
- **Concurrency Engine**: Multi-threading for efficient parallel processing
- **Progress Monitoring**: Real-time progress bars and statistics

### File Structure
- **cmd/**: Cobra-based CLI commands and entrypoints
- **database/**: Database connectors and abstraction layer (MySQL via GORM, Oracle via native SQL)
- **compare/**: SQL comparison logic
- **config/**: Configuration loading and validation  
- **models/**: GORM model definitions for internal metadata
- **pkg/**: Reusable utility packages

## Installation

### Dependencies

1. **External Tools** (for data loading):
   - `sqluldr2` for Oracle data extraction
   - `sqlldr` for Oracle data loading
   - `tidb-lightning` for fast MySQL/TiDB loading
   - `dumpling` for MySQL/TiDB extraction

2. **Go Modules Requirements**:
   - Go 1.19 or higher

### Build

1. Clone the repository:
   ```bash
   git clone https://github.com/Win-Man/dbcompare.git
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build using Makefile:
   ```bash
   make build
   ```

## Configuration

All tools require configuration files specifying:

- Database connection details (host, port, user, password)
- Schema information
- Processing parameters
- Path configurations for temporary files
- Performance settings

Refer to the `.tmpl` files included with the distribution for example configurations.

## Usage Examples

### SQL Comparison
```bash
./bin/dbcompare sql-diff --config config.yaml --sql "SELECT COUNT(*) FROM users"
```

### Data Sync (Prepare phase) 
```bash
./bin/dbcompare sync-diff prepare --config sync_config.yaml
```

### Data Sync (Run phase)  
```bash  
./bin/dbcompare sync-diff run --config sync_config.yaml
```

### Oracle to MySQL/Migration preparation 
```bash
./bin/dbcompare o2t-init all --config o2t_config.yaml
```

### MySQL to Oracle
```bash
./bin/dbcompare t2o-init all --config t2o_config.yaml
```

## Project Status

- Active development tool for database migrations
- Specializes in Oracle↔MySQL/TiDB migration scenarios
- Uses battle-tested components like Dumpling, TiDB Lightning, and SQL*Loader
- Includes sophisticated progress tracking and error handling

## Contributing

- Contributions welcome for additional database connector support
- Enhancement proposals for performance optimizations
- Bug reports and fixes for edge cases in data conversion

## License

Apache 2.0 - See LICENSE file for details

## Notes

- Requires setup of external database tools (sqluldr2, sqlldr, dumpling, tidb-lightning)
- Configuration templates are provided in the bin/ directory after build
- Designed for enterprise database migration scenarios
- Implements robust multi-threading for efficient large-scale conversions

## Roadmap

As indicated in the original TODO list:
- Enhanced logging and completion status reporting
- Additional database type support
- Integration with more external database tools