# CMD/ KNOWLEDGE BASE

**Generated:** 2026-03-01
**Commit:** a5c3e8d

## OVERVIEW
CLI command implementation using Cobra framework.

## STRUCTURE
```
cmd/
├── root.go           # Main cobra command initialization
├── sql_diff.go       # SQL difference analysis command
├── sync_diff_control.go # Sync diff control command
├── o2t_init.go       # Oracle-to-Target initialization
├── t2o_init.go       # Target-to-Oracle initialization
└── ...               # Test files
```

## WHERE TO LOOK
| Task | File | Notes |
|------|------|-------|
| Add new commands | root.go | Register with rootCmd.AddCommand() |
| SQL diff logic | sql_diff.go | Main SQL comparison algorithm |
| Sync control | sync_diff_control.go | Configuration generation logic |

## CONVENTIONS
- All commands use Cobra framework conventions
- Flag definitions follow Cobra standard patterns (VarP)
- Commands registered in root.go via new*Cmd() factory functions

## ANTI-PATTERNS (CMD DIR SPECIFIC)
- Don't modify root.go global vars directly - use cmd.Flags() methods
- Don't create cobra commands outside of factory functions in this directory

## UNIQUE STYLES
- Factory functions for creating commands (newSqlDiffCmd(), etc.)
- Centralized initialization in root.go
- Shared flags for configuration, logging across all commands