#!/usr/bin/env bash
# shellcheck disable=SC2086

# Strict error handling
set -euo pipefail

# Logging and output formatting
readonly SCRIPT_NAME=$(basename "$0")
readonly LOG_FILE="/tmp/${SCRIPT_NAME}.log"

# Color codes for output
readonly COLOR_GREEN='\033[0;32m'
readonly COLOR_RED='\033[0;31m'
readonly COLOR_YELLOW='\033[1;33m'
readonly COLOR_RESET='\033[0m'

# Logging function
log() {
  local level="$1"
  local message="$2"
  local timestamp
  timestamp=$(date "+%Y-%m-%d %H:%M:%S")

  case "${level}" in
    INFO)
      echo -e "${COLOR_GREEN}[INFO] ${timestamp}: ${message}${COLOR_RESET}"
      ;;
    WARN)
      echo -e "${COLOR_YELLOW}[WARN] ${timestamp}: ${message}${COLOR_RESET}" >&2
      ;;
    ERROR)
      echo -e "${COLOR_RED}[ERROR] ${timestamp}: ${message}${COLOR_RESET}" >&2
      ;;
    *)
      echo "[${level}] ${message}"
      ;;
  esac

  echo "[${level}] ${timestamp}: ${message}" >> "${LOG_FILE}"
}

# Error handling
handle_error() {
  log ERROR "Command failed with exit code $?"
  exit 1
}

# Trap errors
trap handle_error ERR

# Default configuration
CONFIG_FILE=""
EXAMPLE_DIR=""

# Load example-specific configuration
load_example_config() {
  local example_name="${1}"

  # Dynamically set paths based on example name
  EXAMPLE_DIR="tests/examples/${example_name}"
  CONFIG_FILE="${EXAMPLE_DIR}/tag-compliance.yaml"

  # Validate configuration file exists
  if [[ ! -f "${CONFIG_FILE}" ]]; then
    log ERROR "Configuration file not found: ${CONFIG_FILE}"
    exit 1
  fi

  log INFO "Loaded configuration for example: ${example_name}"
}

# Main function to run aws-taggy
run_aws_taggy() {
  local example_name="${1}"
  local mode="${2:-check}"
  local output="${3:-table}"

  # Load example-specific configuration
  load_example_config "${example_name}"

  log INFO "Running aws-taggy in ${mode} mode with ${output} output"

  case "${mode}" in
    check)
      aws-taggy compliance check \
        --config "${CONFIG_FILE}" \
        --output="${output}" \
        --detailed
      ;;
    validate)
      aws-taggy compliance validate \
        --config "${CONFIG_FILE}"
      ;;
    report)
      aws-taggy compliance report \
        --config "${CONFIG_FILE}" \
        --output="${output}"
      ;;
    *)
      log ERROR "Invalid mode: ${mode}"
      exit 1
      ;;
  esac
}

# Main script execution
main() {
  local example_name="${1}"
  local mode="${2:-check}"
  local output="${3:-table}"

  if [[ -z "${example_name}" ]]; then
    log ERROR "Example name must be provided"
    exit 1
  fi

  log INFO "Starting aws-taggy script for example: ${example_name}"

  run_aws_taggy "${example_name}" "${mode}" "${output}"

  log INFO "aws-taggy operation completed successfully"
}

# Script entry point
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  main "$@"
fi
