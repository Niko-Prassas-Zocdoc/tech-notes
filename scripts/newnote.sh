#!/bin/bash

# Usage: ./scripts/newnote.sh <category> <slug>
# Example: ./scripts/newnote.sh bugs grpc-timeout

set -euo pipefail

# =======================
# 🔧 Config
# =======================

# Default editor to open the note
EDITOR_CMD="cursor"  # change to "vim", "nano", "subl", etc.

# Allowed categories
VALID_CATEGORIES=("bugs" "infra" "tools" "scripts" "notes")

# =======================
# 🚦 Input check
# =======================

if [ "$#" -lt 2 ]; then
  echo "Usage: $0 <category> <slug>"
  echo "Example: $0 bugs grpc-timeout"
  exit 1
fi

category=$1
slug=$2

# Validate category
if [[ ! " ${VALID_CATEGORIES[@]} " =~ " ${category} " ]]; then
  echo "❌ Invalid category: '$category'"
  echo "Allowed categories are: ${VALID_CATEGORIES[*]}"
  exit 1
fi

# =======================
# 🛠️ Paths
# =======================

base_dir=$(dirname "$(dirname "$0")") # Go to repo root
folder="$base_dir/$category"
date=$(date +%Y-%m-%d)
filename="$folder/${date}-${slug}.md"
template="$base_dir/template.md"

# Check template exists
if [ ! -f "$template" ]; then
  echo "❌ Template not found at: $template"
  echo "Make sure 'template.md' exists in the repo root."
  exit 1
fi

# =======================
# 📝 Create note
# =======================

mkdir -p "$folder"
cp "$template" "$filename"

echo "✅ Note created at: $filename"

# =======================
# 🖊️ Open in editor
# =======================

if command -v "$EDITOR_CMD" &> /dev/null; then
  "$EDITOR_CMD" "$filename"
else
  echo "⚠️ Editor command '$EDITOR_CMD' not found. Open the file manually:"
  echo "$filename"
fi
