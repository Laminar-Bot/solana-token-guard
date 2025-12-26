#!/bin/bash
# Block edits to sensitive files

input=$(cat)
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty')

# Block sensitive files
if [[ "$file_path" =~ \.(env|env\.local|secrets|private)$ ]]; then
  cat <<EOF
{
  "decision": "block",
  "message": "🚫 Cannot edit sensitive file: $file_path\n\nUse AWS Secrets Manager or manual editing with proper secret management."
}
EOF
  exit 1
fi

# Block package-lock.json
if [[ "$file_path" == "package-lock.json" ]]; then
  cat <<EOF
{
  "decision": "block",
  "message": "🚫 Do not edit package-lock.json directly.\n\nUse 'npm install <package>' instead."
}
EOF
  exit 1
fi

exit 0
