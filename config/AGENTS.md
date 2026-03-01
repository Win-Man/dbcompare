# CONFIG/ KNOWLEDGE BASE

**Generated:** 2026-03-01
**Commit:** a5c3e8d

## OVERVIEW
Configuration management and parsing for YAML and TOML formats.

## WHERE TO LOOK
| Task | File | Notes |
|------|------|-------|
| YAML config | config.go | YAML configuration parsing, validation |
| TOML config | oto_config.go | TOML format configuration |
| Base config | config.go | Common configuration structures |
 
## CONVENTIONS
- Support both YAML and TOML format configurations
- Config struct hierarchy mirrors file structure
- Validation occurs during parsing phase

## ANTI-PATTERNS (CONFIG DIR SPECIFIC)
- Don't add environment-specific values directly to config structs
- Don't mix runtime parameters with static configuration settings

## UNIQUE STYLES
- Dual format support for configuration files (YAML/TOML)
- Template-based config generation 
- Centralized config structures accessible throughout app