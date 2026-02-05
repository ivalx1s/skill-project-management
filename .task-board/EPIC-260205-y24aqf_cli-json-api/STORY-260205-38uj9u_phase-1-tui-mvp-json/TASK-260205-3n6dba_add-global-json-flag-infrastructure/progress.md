## Status
done

## Assigned To
agent-json-infra

## Created
2026-02-05T11:17:01Z

## Last Update
2026-02-05T11:21:13Z

## Blocked By
- (none)

## Blocks
- TASK-260205-ohoqtl
- TASK-260205-1othrv
- TASK-260205-rl3fap

## Checklist
(empty)

## Notes
Implemented global --json flag infrastructure:
- Added global --json flag to root.go (PersistentFlags)
- Created internal/output/json.go with JSON output helpers
- Defined error codes: NOT_FOUND, INVALID_ID, INVALID_STATUS, CYCLE_DETECTED, VALIDATION_ERROR, INTERNAL_ERROR
- Added helper functions: PrintJSON, PrintError, NewErrorResponse, MarshalJSON
- Added comprehensive tests in json_test.go
- All tests pass, code linted
