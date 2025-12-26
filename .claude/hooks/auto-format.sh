#!/bin/bash
# Auto-format TypeScript files after edits

input=$(cat)
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty')

# Only format TypeScript/JavaScript
if [[ ! "$file_path" =~ \.(ts|tsx|js|jsx)$ ]]; then
  exit 0
fi

# Run prettier
if command -v npx &> /dev/null && npx prettier --write "$file_path" 2>/dev/null; then
  echo "✅ Formatted: $file_path" >&2
  exit 0
fi

exit 0
