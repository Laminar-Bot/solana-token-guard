#!/bin/bash
# Auto-sync OpenAPI spec when controller files change

input=$(cat)
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty')

# Only trigger on controller file changes
if [[ ! "$file_path" =~ \.controller\.ts$ ]]; then
  exit 0
fi

# Check if fastify-swagger is available (optional dependency)
if ! command -v npx &> /dev/null; then
  exit 0
fi

# Regenerate OpenAPI spec (if script exists)
if [ -f "scripts/generate-openapi-spec.ts" ]; then
  npx tsx scripts/generate-openapi-spec.ts 2>/dev/null

  if [ $? -eq 0 ]; then
    echo "✅ OpenAPI spec updated: docs/03-TECHNICAL/architecture/api-specification.md" >&2
  fi
fi

exit 0
