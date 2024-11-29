#!/usr/bin/env bash
# shellcheck disable=SC1091,SC2086

# Determine the project root directory
PROJECT_ROOT=$(git rev-parse --show-toplevel)

# Source utility scripts
source "${PROJECT_ROOT}/scripts/run_me.sh"
source "${PROJECT_ROOT}/scripts/terraform_manage.sh"

# Example-specific configuration
readonly EXAMPLE_NAME="example-s3-specific-tags"

# Validate AWS credentials
validate_aws_credentials() {
  # Check for required AWS environment variables
  local required_vars=("AWS_ACCESS_KEY_ID" "AWS_SECRET_ACCESS_KEY" "AWS_REGION")
  local missing_vars=()

  for var in "${required_vars[@]}"; do
    if [[ -z "${!var}" ]]; then
      missing_vars+=("$var")
    fi
  done

  # If any required variables are missing, exit with an error
  if [[ ${#missing_vars[@]} -gt 0 ]]; then
    log ERROR "Missing AWS credentials. Please export the following variables:"
    for missing_var in "${missing_vars[@]}"; do
      log ERROR "  export ${missing_var}=your_value"
    done
    exit 1
  fi

  # Additional validation: check if AWS CLI can be used
  if ! aws sts get-caller-identity &>/dev/null; then
    log ERROR "AWS credentials validation failed. Unable to authenticate with AWS."
    log ERROR "Please check your AWS credentials and try again."
    exit 1
  fi

  log INFO "AWS credentials validated successfully"
}

# Wrapper function to run the example
run_example() {
  local mode="${1:-all}"

  # Validate AWS credentials before any operation
  validate_aws_credentials

  # Validate input mode
  case "${mode}" in
    terraform)
      # Only run Terraform operations
      manage_terraform "${EXAMPLE_NAME}" "apply"
      ;;
    compliance)
      # Only run compliance check
      run_aws_taggy "${EXAMPLE_NAME}"
      ;;
    destroy)
      # Destroy resources
      manage_terraform "${EXAMPLE_NAME}" "destroy"
      ;;
    all)
      # Full workflow: apply, check compliance, then destroy
      manage_terraform "${EXAMPLE_NAME}" "apply"
      run_aws_taggy "${EXAMPLE_NAME}"
      manage_terraform "${EXAMPLE_NAME}" "destroy"
      ;;
    *)
      # Invalid mode
      log ERROR "Invalid mode: ${mode}. Supported modes: terraform, compliance, destroy, all"
      exit 1
      ;;
  esac
}

# Main script execution
main() {
  local mode="${1:-all}"

  # Log the start of the example
  log INFO "Running S3 tag compliance example with mode: ${mode}"

  # Run the example
  run_example "${mode}"

  # Log successful completion
  log INFO "S3 tag compliance example completed successfully"
}

# Script entry point
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  main "$@"
fi
