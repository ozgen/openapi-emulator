# openapi-sample-emulator

`openapi-sample-emulator` is a small HTTP emulator that returns predefined JSON responses based on an OpenAPI / Swagger specification.

It is mainly intended for local development, integration testing, and CI environments where stable and predictable API responses are required.

---

## What it does

* Reads a Swagger / OpenAPI specification
* Matches incoming requests by method and path
* Returns responses from JSON sample files
* Optionally falls back to examples defined in the spec
* Can enforce simple request validation (e.g. required body)

---

## Why use it

This tool is useful when:

* You need deterministic responses (no random data)
* You want to test integrations without running real services
* Your API spec is Swagger 2.0 and lacks good examples
* CI tests must be repeatable and stable

---

## How responses are chosen

1. If a matching sample file exists, it is returned
2. If not, response examples from the spec are used (if available)
3. Otherwise, a minimal response is generated from the schema

---

## Sample files

Sample files are matched using this naming rule:

```
METHOD__path_with_slashes_replaced_by_underscores.json
```

Examples:

* `GET /api/v1/items` -> `GET__api_v1_items.json`
* `POST /scans` -> `POST__scans.json`
* `GET /scans/{id}/results` -> `GET__scans_{id}_results.json`

Path parameters stay as `{id}`.

---

## Running with Docker

```bash
docker compose up -d
```
or build and run with docker:

```bash
 docker compose up --build
```

Minimal `docker-compose.yml` example:

```yaml
services:
  emulator:
    image: openapi-sample-emulator:local
    ports:
      - "8086:8086"
    environment:
      - SERVER_PORT=8086
      - SPEC_PATH=/work/swagger.json
      - SAMPLES_DIR=/work/sample
      - FALLBACK_MODE=openapi_examples
      - VALIDATION_MODE=required
    volumes:
      - ./examples/demo:/work:ro
```

---

## Validation

Optional request validation can be enabled:

```bash
VALIDATION_MODE=required
```

Currently supported:

* Required request body (Swagger 2.0 `required: true` body parameters)

---

## When not to use it

This tool is not intended to:

* Generate random mock data
* Fully validate request schemas
* Replace contract-testing tools

---

## License

MIT

---
