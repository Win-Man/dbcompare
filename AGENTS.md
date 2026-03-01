# PROJECT KNOWLEDGE BASE

**Generated:** 2026-03-01
**Commit:** a5c3e8d
**Branch:** main

## OVERVIEW
Database comparison tool supporting MySQL & Oracle. CLI app with multiple subcommands using Cobra, backed with GORM for database operations.

## STRUCTURE
```
./
├── cmd/          # CLI commands (sql-diff, sync-diff-control, o2t-init, t2o-init)
├── database/     # Database drivers (MySQL, Oracle)
├── models/       # GORM data models  
├── config/       # Configuration management (YAML, TOML)
├── pkg/          # Utility packages (progress, remote execution)
├── service/      # Core business logic
└── dev/conf/     # Development configurations
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| CLI commands | cmd/ | Add new commands in root.go |
| DB drivers | database/ | GORM wrappers for MySQL/Oracle |
| Data models | models/ | GORM struct definitions |
| Configuration | config/ | YAML/TOML parsing logic |

## CONVENTIONS
- Go 1.19 project
- Uses Cobra for CLI framework
- GORM for database operations 
- No external test framework (stdlib testing only)
- Build via Makefile with version embedding

## ANTI-PATTERNS (THIS PROJECT)
- Don't use both pkg/log and pkg/logger - consolidate to one
- Don't create new main() outside of cmd/ - should integrate with Cobra
- Invalid go.mod replace directives shouldn't reference nonexistent paths

## UNIQUE STYLES
- Heavy use of ldflags for version embedding in build
- Database ops abstracted via GORM layer
- CLI flags managed through Cobra

## COMMANDS
```bash
# Build binary
make build

# Format code
make tool

# Run linter
make lint

# Cross-compile
make arm64 amd64
```

## NOTES
- Contains development configs in dev/ that may be for internal use
- Invalid replace directives in go.mod need cleanup
- Has standalone roundt.go in usua/ that isn't connected to main CLI