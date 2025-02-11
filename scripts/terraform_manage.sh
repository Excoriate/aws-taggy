#!/usr/bin/env bash
# shellcheck disable=SC2086

# Strict error handling
set -euo pipefail

# Logging and output formatting
# Use a function to get script name that works when sourced
get_script_name() {
  local script_path="${BASH_SOURCE[0]}"
  if [[ -z "$script_path" ]]; then
    script_path="$0"
  fi
  basename "$script_path"
}

# Prevent variable reassignment
if [ -z "${SCRIPT_NAME+x}" ]; then
  SCRIPT_NAME=$(get_script_name)
  LOG_FILE="/tmp/${SCRIPT_NAME}.log"

  # Color codes for output
  COLOR_GREEN='\033[0;32m'
  COLOR_RED='\033[0;31m'
  COLOR_YELLOW='\033[1;33m'
  COLOR_RESET='\033[0m'
fi

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
EXAMPLE_DIR=""

# Load example-specific configuration
load_example_config() {
  local example_name="${1}"
  local example_dir_candidates=(
    "tests/examples/${example_name}"
    "tests/examples/example-${example_name}"
  )

  # Try different directory paths
  for dir in "${example_dir_candidates[@]}"; do
    if [[ -f "${dir}/main.tf" ]]; then
      EXAMPLE_DIR="${dir}"
      log INFO "Loaded Terraform configuration for example: ${example_name}"
      return
    fi
  done

  # If no configuration found, exit with an error
  log ERROR "Terraform configuration not found for example: ${example_name}"
  log ERROR "Tried directories: ${example_dir_candidates[*]}"
  exit 1
}

# Initialize Terraform
init_terraform() {
  local example_name="${1}"

  log INFO "Initializing Terraform for ${example_name}"
  terraform -chdir="${EXAMPLE_DIR}" init \
    -upgrade=true \
    -input=false
}

# Validate Terraform configuration
validate_terraform() {
  local example_name="${1}"

  log INFO "Validating Terraform configuration for ${example_name}"
  terraform -chdir="${EXAMPLE_DIR}" validate
}

# Plan Terraform changes
plan_terraform() {
  local example_name="${1}"

  log INFO "Planning Terraform changes for ${example_name}"
  terraform -chdir="${EXAMPLE_DIR}" plan \
    -out=tfplan \
    -input=false
}

# Apply Terraform changes
apply_terraform() {
  local example_name="${1}"

  log INFO "Applying Terraform changes for ${example_name}"
  terraform -chdir="${EXAMPLE_DIR}" apply \
    -auto-approve \
    -input=false \
    tfplan
}

# Destroy Terraform resources
destroy_terraform() {
  local example_name="${1}"

  log INFO "Destroying Terraform resources for ${example_name}"
  terraform -chdir="${EXAMPLE_DIR}" destroy \
    -auto-approve \
    -input=false
}

# Main function to manage Terraform operations
manage_terraform() {
  local example_name="${1}"
  local mode="${2:-plan}"

  # Load example-specific configuration
  load_example_config "${example_name}"

  # Initialize and validate
  init_terraform "${example_name}"
  validate_terraform "${example_name}"

  case "${mode}" in
    plan)
      plan_terraform "${example_name}"
      ;;
    apply)
      plan_terraform "${example_name}"
      apply_terraform "${example_name}"
      ;;
    destroy)
      destroy_terraform "${example_name}"
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
  local mode="${2:-plan}"

  if [[ -z "${example_name}" ]]; then
    log ERROR "Example name must be provided"
    exit 1
  fi

  log INFO "Starting Terraform management script for example: ${example_name}"

  manage_terraform "${example_name}" "${mode}"

  log INFO "Terraform operation completed successfully"
}

# Script entry point
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  main "$@"
fi
