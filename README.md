# alelm-gocore

Reusable Golang core library for scalable backend applications.

---

# Overview

`alelm-gocore` adalah reusable backend foundation package berbasis Golang.

Tujuan utama:
- mempercepat development
- mengurangi duplicate code
- menyediakan reusable utilities
- menyediakan scalable backend foundation
- mendukung clean architecture

Package ini dirancang untuk:
- monolith
- modular monolith
- microservices
- SaaS applications
- ERP systems
- booking systems
- fintech systems

---

# Philosophy

`alelm-gocore` fokus pada:

- reusable engineering
- modularity
- maintainability
- scalable architecture
- clean code
- developer productivity

---

# Main Features

## Generic CRUD

Features:
- generic CRUD
- pagination
- filtering
- sorting
- soft delete
- relation support

---

## Dynamic Search

Features:
- relation search
- dynamic joins
- reusable query builder
- search tag parser

---

## Reflection System

Features:
- metadata cache
- struct parser
- field extraction
- reusable reflection utilities

---

## Database Utilities

Features:
- PostgreSQL helper
- transaction helper
- query abstraction
- migration helper

---

## Response Utilities

Features:
- standardized API response
- pagination response
- error response

---

## Error Handling

Features:
- application errors
- validation errors
- HTTP error mapping

---

## Authentication Utilities

Features:
- JWT helper
- token parser
- auth middleware
- permission helper

---

## Middleware

Features:
- auth middleware
- logger middleware
- recovery middleware
- request middleware

---

## Validation

Features:
- reusable validators
- request validation
- custom validation rules

---

## Logger

Features:
- structured logging
- zap wrapper
- request logger

---

## Config

Features:
- env loader
- config parser
- environment helper

---

## Cache

Features:
- memory cache
- redis helper
- cache abstraction

---

## Indonesia Utilities

Features:
- province data
- regency data
- district data
- village data
- postal code utilities

---

## Utility Helpers

Features:
- string helper
- slice helper
- time helper
- pointer helper

---

# Package Structure

```txt
alelm-gocore/
│
├── auth/
├── cache/
├── config/
├── crud/
├── database/
├── errors/
├── filters/
├── indonesia/
├── logger/
├── middleware/
├── pagination/
├── reflection/
├── response/
├── search/
├── utils/
├── validator/
└── examples/
```

---

# Recommended Usage

Recommended architecture:

```txt
handler
  ↓
usecase/service
  ↓
repository/accessor
  ↓
database
```

---

# Goals

## Reusable

Can be used across multiple projects.

---

## Modular

Each package should work independently.

---

## Framework Agnostic

Can be used with:
- Fiber
- Gin
- Echo
- net/http

---

## Database Agnostic

Should support:
- PostgreSQL
- MySQL
- SQLite

---

# Example Use Cases

- booking systems
- ERP systems
- fintech systems
- school systems
- sports platforms
- SaaS applications

---

# Future Plans

Planned future modules:
- OpenAPI helper
- websocket helper
- queue helper
- event system
- distributed cache
- RBAC system
- audit log system

---

# AI Context

This repository is intended to:
- reduce repetitive backend code
- provide reusable architecture
- standardize backend development
- improve developer productivity

When generating code:
- prioritize reusability
- prioritize modularity
- prioritize clean architecture
- avoid unnecessary abstraction
- avoid overengineering
