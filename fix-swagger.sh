#!/bin/bash
echo "=== Fixing swagger host ==="
find . -name "swagger.json" -exec sed -i 's/localhost:5685/24.144.84.159/g' {} \;
find . -name "*.json" -exec grep -l "localhost:5685" {} \; || echo "No localhost found"
echo "=== Swagger fix complete ==="
