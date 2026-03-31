## Overview

This API was designed to support a file processing pipeline that ingests, transforms, and updates member data at scale.

It provides endpoints for:
- retrieving member records
- updating and writing processed data
- supporting reliable data workflows from external processors

This service is used in conjunction with the high-performance file processor:
https://github.com/NTRCodes/high-performance-file-processor

# Member Management API

> **Note:** This is a sanitized version of a production API built for a membership organization.
> Configuration details, company names, and IP addresses have been removed or generalized.

**Production Impact:**
- ✅ Serving 1,000+ requests/minute in production
- ✅ Sub-100ms response times for member lookups
- ✅ 86% reduction in Docker image size (526MB → 72MB)
- ✅ Deployed with CI/CD pipeline and zero-downtime updates

---

A **production-ready Go API** for member management and metrics tracking.

## 🚀 Features

- **Fast & Lightweight** - 72MB Docker image
- **Complete CRUD API** - Full member management
- **Metrics Tracking** - Performance and ROI metrics
- **Secure** - API key authentication
- **Containerized** - Multi-stage Docker build
- **Production Tested** - 99.9% uptime

## 📚 API Overview

### Member Management
- `POST /members` - Create new member
- `GET /members/{id}` - Get member by ID
- `PUT /members/{id}` - Update member
- `DELETE /members/{id}` - Delete member

### Metrics
- `POST /metrics/performance` - Record performance metrics
- `POST /metrics/value` - Record cost savings
- `POST /metrics/batch-jobs` - Record batch processing

### System
- `GET /healthz` - Health check
- `GET /redoc` - API documentation

## 🚀 Quick Start

```bash
# Start with Docker Compose
docker compose up --build

# API available at http://localhost:8000
```

## 🧪 Testing

```bash
go test ./... -coverprofile=coverage.out -v
go tool cover -func=coverage.out
```

## 📊 Performance

- **Member Lookup:** < 50ms (p95)
- **Metrics Insert:** < 30ms (p95)
- **Uptime:** 99.9%

**Tech:** Go · PostgreSQL · Docker · REST API
