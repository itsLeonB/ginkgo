# Error Handling Convention

This document describes the error handling convention expected by the error middleware. All developers contributing handlers or services must follow these rules to ensure consistent, observable, and correctly classified API error responses.

---

## Overview

The middleware intercepts all errors attached to the Gin context and maps them to structured JSON responses. It distinguishes between several error categories, each with different logging severity and HTTP response behavior. The `ungerr` package is the primary tool for signaling these categories.

---

## Error Categories

### 1. Expected Application Errors — `ungerr.AppError`

Use this for **known, anticipated client-side failure conditions** — these always map to `4xx` responses. Examples include unauthorized access, missing resources, or input that fails business rule validation.

```go
// In a handler or service
if !authorized {
    return ungerr.ForbiddenError("you do not have access to this resource")
}

if user == nil {
    return ungerr.NotFoundError("user not found")
}
```

**What the middleware does:** logs at `WARN` and responds with the AppError's HTTP status and detail.

---

### 2. Wrapping Unexpected External Errors — `ungerr.Wrap()`

Use this when you call an **external dependency** (database, HTTP client, file system, third-party or standard library, etc.) and get back an error you didn't anticipate. This excludes internal libraries that already use `ungerr` for their errors — those will surface as `AppError` and be handled automatically.

```go
user, err := repo.FindByID(id)
if err != nil {
    return ungerr.Wrap(err, "failed to find user by id")
}
```

**What the middleware does:**
- Attempts to identify the wrapped cause (e.g. validation errors, JSON errors, EOF, broken pipe).
- If identifiable → logs at `WARN` as `"identified wrapped error"`, responds with the appropriate AppError status.
- If not identifiable → logs at `ERROR` as `"unhandled error"`, responds with `500 Internal Server Error`.

> Wrapping preserves the original error for logging while keeping the response safe for clients.

---

### 3. Unreachable / Defensive Guards — `ungerr.Unknown()`

Use this for **code paths that should never be reached** — defensive guards, exhaustive switch statements, or post-condition assertions. Pass a message describing what assumption was violated.

```go
switch status {
case "active", "inactive":
    // handled
default:
    // This should never happen if the DB has a proper constraint
    return ungerr.Unknown("unexpected user status: " + status)
}
```

**What the middleware does:** logs at `ERROR` as `"unexpected error"`, responds with `500 Internal Server Error`.

> Developers should keep an eye on these in logs — they indicate a broken assumption in the code that needs to be investigated and fixed.

---

## What NOT to Do

### ❌ Returning raw errors without wrapping

```go
// BAD — do not do this
user, err := repo.FindByID(id)
if err != nil {
    return err
}
```

The middleware will catch this, but it logs at `ERROR` with a message explicitly telling developers the error was not wrapped:

```text
unwrapped error detected — wrap with ungerr.Wrap()
```

It will still return `500` to the client, but this is treated as a developer mistake, not a handled case. Always wrap external errors.

---

## Decision Guide

Use this to decide which tool to reach for:

```text
Did something go wrong?
│
├── Is it an expected client-side failure? (4xx)
│   └── Yes → ungerr.SomeNamedError()                    (AppError)
│
├── Did an external call (DB, HTTP, stdlib, third-party lib) return an unexpected error?
│   └── Yes → ungerr.Wrap(err, "context message")        (UnknownError with cause)
│
└── Did execution reach a place it logically never should?
    └── Yes → ungerr.Unknown("what assumption was broken") (UnknownError, no cause)
```

---

## Log Level Reference

| Situation | Log Level | Log Message |
|---|---|---|
| Expected AppError returned | `WARN` | `"application error"` |
| Wrapped error, cause identifiable | `WARN` | `"identified wrapped error"` |
| Wrapped error, cause unidentifiable | `ERROR` | `"unhandled error"` |
| `Unknown(nil)` — no cause | `ERROR` | `"unexpected error"` |
| Raw error, not wrapped at all | `ERROR` | `"unwrapped error detected — wrap with ungerr.Wrap()"` |
| Panic recovered | `ERROR` | `"panic recovered"` |

---

## Automatically Identified Error Types

When using `ungerr.Wrap(err)`, the middleware can automatically identify and map the following cause types to appropriate HTTP responses, without any extra code from you:

| Cause type | HTTP Response |
|---|---|
| `validator.ValidationErrors` | `422 Unprocessable Entity` |
| `*json.SyntaxError` | `400 Bad Request` — invalid JSON |
| `*json.UnmarshalTypeError` | `400 Bad Request` — invalid field value |
| `io.EOF` | `400 Bad Request` — missing request body |
| `"connection reset by peer"` | `400 Bad Request` — connection error |
| `"broken pipe"` | `400 Bad Request` — connection error |

Any other cause falls through to `500 Internal Server Error`.
