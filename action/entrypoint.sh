#!/bin/sh
set -e

# Parse the validate-only flag to determine which command to run
VALIDATE_ONLY="false"
COMMAND_ARGS=""

# Process arguments to extract validate-only flag
while [ $# -gt 0 ]; do
  case $1 in
    --validate-only)
      VALIDATE_ONLY="$2"
      shift 2
      ;;
    --client-id|--client-secret)
      # Only add credentials for process command
      if [ "$VALIDATE_ONLY" != "true" ]; then
        COMMAND_ARGS="$COMMAND_ARGS $1 $2"
      fi
      shift 2
      ;;
    *)
      COMMAND_ARGS="$COMMAND_ARGS $1"
      shift
      ;;
  esac
done

# Determine which command to run based on validate-only flag
BINARY_PATH="/app/nobl9-action"

# For local testing, use current directory binary if /app doesn't exist
if [ ! -f "$BINARY_PATH" ]; then
  BINARY_PATH="./nobl9-action"
fi

if [ "$VALIDATE_ONLY" = "true" ]; then
  echo "Running validation mode..."
  exec $BINARY_PATH validate $COMMAND_ARGS
else
  echo "Running process mode..."
  exec $BINARY_PATH process $COMMAND_ARGS
fi