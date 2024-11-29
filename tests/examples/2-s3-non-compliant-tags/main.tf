provider "aws" {
  region = "us-east-1"
}

resource "aws_s3_bucket" "non_compliant_bucket" {
  bucket        = "aws-taggy"
  force_destroy = true

  tags = {
    # Prohibited tags
    "temp:test"     = "temporary-resource"
    "test:example"  = "should-not-be-here"

    # Invalid tag keys
    "ENVIRONMENT"   = "PRODUCTION"  # Uppercase, not allowed
    "Owner"         = "john.doe@gmail.com"  # Wrong email domain

    # Incorrect required tags
    "project"       = "Invalid-Project-Name"  # Contains uppercase
    "environment"   = "development"  # Not in allowed values
    "data_class"    = "unknown"  # Not in allowed values
    "cost_center"   = "1234"  # Wrong format

    # Additional problematic tags
    "Sensitive"     = "High Risk"
    "random:tag"    = "unexpected value"
  }
}

# Optional: Add some bucket configurations
resource "aws_s3_bucket_versioning" "versioning" {
  bucket = aws_s3_bucket.non_compliant_bucket.id
  versioning_configuration {
    status = "Enabled"
  }
}
