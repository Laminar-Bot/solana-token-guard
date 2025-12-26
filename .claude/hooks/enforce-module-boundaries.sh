#!/bin/bash
# Detect illegal module imports (must use index.ts public API)

input=$(cat)
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty')
new_string=$(echo "$input" | jq -r '.tool_input.new_string // empty')

# Only check module files
if [[ ! "$file_path" =~ src/modules/ ]]; then
  exit 0
fi

# Check for illegal direct imports (bypass index.ts)
if echo "$new_string" | grep -qE "from ['\"]\.\./[^/]+/[^index]"; then
  cat <<EOF
{
  "decision": "block",
  "message": "🚫 Module boundary violation detected.\n\nDirect imports to module internals are forbidden.\n\n❌ BAD: import { X } from '../scan/scan.service'\n✅ GOOD: import { X } from '../scan' (uses public API)\n\nSee: .claude/agents/modular-monolith-architect.md"
}
EOF
  exit 1
fi

exit 0
