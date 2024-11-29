#!/usr/bin/env bash
# shellcheck disable=SC1091,SC2086

# Determine the project root directory
PROJECT_ROOT=$(git rev-parse --show-toplevel)

# Source utility scripts
source "${PROJECT_ROOT}/scripts/run_me.sh"
source "${PROJECT_ROOT}/scripts/terraform_manage.sh"

# Example-specific configuration
EXAMPLE_NAME="1-s3-specific-tags"

# Validate AWS credentials
validate_aws_credentials() {
  # Check for required AWS environment variables
  local required_vars=("AWS_ACCESS_KEY_ID" "AWS_SECRET_ACCESS_KEY")
  local missing_vars=()

  for var in "${required_vars[@]}"; do
    if [[ -z "${!var+x}" ]]; then
      missing_vars+=("$var")
    fi
  done

  # Default AWS_REGION to us-east-1 if not set
  if [[ -z "${AWS_REGION+x}" ]]; then
    export AWS_REGION="us-east-1"
    log INFO "AWS_REGION not set. Defaulting to us-east-1"
  fi

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
  local mode="${1:-create}"

  # Validate AWS credentials before any operation
  validate_aws_credentials

  # Validate input mode
  case "${mode}" in
    create)
      # Run Terraform create/apply operation
      manage_terraform "${EXAMPLE_NAME}" "apply"
      ;;
    plan)
      # Only run Terraform plan
      manage_terraform "${EXAMPLE_NAME}" "plan"
      ;;
    destroy)
      # Destroy resources
      manage_terraform "${EXAMPLE_NAME}" "destroy"
      ;;
    run)
      # Full scenario: create resources, run compliance check
      manage_terraform "${EXAMPLE_NAME}" "apply"
      run_compliance_check "${EXAMPLE_NAME}"
      ;;
    run-cli)
      # Run compliance check assuming infrastructure is already created
      run_compliance_check "${EXAMPLE_NAME}"
      ;;
    *)
      # Invalid mode
      log ERROR "Invalid mode: ${mode}. Supported modes: create, plan, destroy, run, run-cli"
      exit 1
      ;;
  esac
}

# Run compliance check using CLI from source code
run_compliance_check() {
  local example_name="${1}"
  local config_file="${PROJECT_ROOT}/tests/examples/${example_name}/tag-compliance.yaml"
  local resource_name="aws-taggy"

  log INFO "Running compliance check from source code"
  go run "${PROJECT_ROOT}/cli/main.go" compliance check \
    --config "${config_file}" \
    --resource "${resource_name}" \
    --output=table \
    --detailed
}

# Main script execution
main() {
  local mode="${1:-create}"

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
