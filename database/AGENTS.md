# DATABASE/ KNOWLEDGE BASE

**Generated:** 2026-03-01
**Commit:** a5c3e8d

## OVERVIEW
Database abstraction layer supporting MySQL and Oracle via GORM interface.

## STRUCTURE
```
database/
├── db.go             # Generic database connector interface
├── gorm_mysql.go     # MySQL-specific GORM implementation
├── oracle.go         # Oracle-specific implementation  
├── oracle_test.go    # Oracle connection tests
├── gorm_mysql_test.go # MySQL GORM tests
└── db_test.go       # Base database tests
```

## WHERE TO LOOK
| Task | File | Notes |
|------|------|-------|
| Add new DB providers | db.go | Extend Connector interface |
| MySQL implementation | gorm_mysql.go | GORM MySQL driver specific code |
| Oracle implementation | oracle.go | Oracle-specific connection/query |
| Provider registration | db.go | ConnectFactory pattern |

## CONVENTIONS
- Use GORM for MySQL operations via GormDBWrapper
- Direct Oracle implementation via sql.DB for Oracle
- Connector interface for consistent abstraction
- Connection pooling handled via database/sql

## ANTI-PATTERNS (DATABASE DIR SPECIFIC)
- Don't use raw SQL for MySQL - use GORM ORM layer
- Don't bypass ConnectFactory - always go through provider registry  
- Don't hardcode connection strings in production code

## UNIQUE STYLES
- Abstract Connector interface with provider factory pattern
- Mixed approaches: GORM for MySQL, raw Oracle SQL
- Extensible provider system allows adding new database types