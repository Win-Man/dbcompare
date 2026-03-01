# MODELS/ KNOWLEDGE BASE

**Generated:** 2026-03-01
**Commit:** a5c3e8d

## OVERVIEW
GORM data models defining database schema and relationships.

## WHERE TO LOOK
| Task | File | Notes |
|------|------|-------|
| Table schemas | All files | GORM struct tags define database schema |
| Model relationships | model.go | Foreign keys, associations defined |
| Migration setup | model_test.go | Table creation, sample data |

## CONVENTIONS
- All structs use GORM standard naming conventions
- Struct tags define column mappings, constraints
- Models are tied to GORM-specific operations and callbacks

## UNIQUE STYLES
- Pure GORM models without validation methods
- Minimal business logic on models
- Simple struct definitions focused on schema mapping 