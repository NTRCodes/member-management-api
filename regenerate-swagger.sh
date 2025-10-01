#!/bin/bash
echo "Regenerating Swagger docs with stable version..."
rm -rf docs/
swag init -g cmd/api/main.go
echo "✅ Swagger docs regenerated successfully!"
