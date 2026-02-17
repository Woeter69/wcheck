#!/bin/bash

# .git/hooks/pre-push
# This script runs wcheck scan before every git push.
# If wcheck fails (exit code 1), the push will be canceled.

TARGET_URL="http://localhost:3000"
WORKERS=5

echo "--------------------------------------------------------"
echo "  [wcheck] Scouting $TARGET_URL before push..."
echo "--------------------------------------------------------"

# Run wcheck
wcheck scan "$TARGET_URL" -w "$WORKERS"

# Store the exit code
RESULT=$?

if [ $RESULT -ne 0 ]; then
  echo ""
  echo "  [!] ERROR: wcheck found errors on $TARGET_URL."
  echo "  [!] Push aborted. Please fix errors before pushing."
  exit 1
else
  echo ""
  echo "  [OK] wcheck scouted successfully. Proceeding with push..."
  exit 0
fi
