# Environment Variables

This document lists the environment variables used by **openapi-sample-emulator** and their purpose.  
Defaults are shown in parentheses.

---

## Core Configuration

| Variable           | Default              | Description |
|-------------------|----------------------|-------------|
| `SERVER_PORT`      | `8086`               | Port the emulator listens on. |
| `SPEC_PATH`        | `/work/swagger.json` | Path to the OpenAPI/Swagger spec file (JSON). |
| `SAMPLES_DIR`      | `/work/sample`       | Directory containing JSON sample response files. |
| `LOG_LEVEL`        | `info`               | Logging level (`debug`, `info`, `warn`, `error`). |
| `RUNNING_ENV`      | `docker`             | Runtime environment (`docker`, `k8s`, `local`). |
| `VALIDATION_MODE`  | `required`           | Request validation mode (`none`, `required`). |
| `FALLBACK_MODE`    | `openapi_examples`   | Fallback behavior if a sample file is missing (`none`, `openapi_examples`). |
| `DEBUG_ROUTES`     | `false`              | If `true`, prints the resolved route→sample mappings on startup. |

---

## Behavior Notes

### `FALLBACK_MODE`

When a sample file is missing:

- `openapi_examples`: try to return examples from the spec response (and if none exist, the emulator may generate a minimal JSON from the response schema).
- `none`: return an error response (HTTP 501) that includes the expected sample file name.

### `VALIDATION_MODE`

- `required`: if the spec marks a request body as required, the emulator rejects requests with an empty body (HTTP 400).
- `none`: skips request body presence checks.

### `DEBUG_ROUTES`

When enabled, the emulator prints lines like:

```
GET /items/{id} -> GET__items_{id}.json
POST /items -> POST__items.json
```

---

## Sample `.env`

```env
# Server
SERVER_PORT=8086
LOG_LEVEL=info
RUNNING_ENV=docker

# Spec + Samples
SPEC_PATH=/work/swagger.json
SAMPLES_DIR=/work/sample

# Fallback / Validation
FALLBACK_MODE=openapi_examples   # none | openapi_examples
VALIDATION_MODE=required         # none | required

# Debug
DEBUG_ROUTES=false
````

---
