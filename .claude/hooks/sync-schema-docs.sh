#!/bin/bash
# Auto-sync database schema documentation when Prisma schema changes

input=$(cat)
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty')

# Only trigger on Prisma schema file changes
if [[ "$file_path" != "prisma/schema.prisma" ]]; then
  exit 0
fi

# Check if Prisma is available
if ! command -v npx &> /dev/null; then
  exit 0
fi

# Regenerate Prisma client
npx prisma generate 2>/dev/null

# Sync schema docs (if script exists)
if [ -f "scripts/sync-schema-docs.ts" ]; then
  npx tsx scripts/sync-schema-docs.ts 2>/dev/null

  if [ $? -eq 0 ]; then
    echo "✅ Database schema documentation updated: docs/03-TECHNICAL/architecture/data-model.md" >&2
  fi
fi

exit 0
