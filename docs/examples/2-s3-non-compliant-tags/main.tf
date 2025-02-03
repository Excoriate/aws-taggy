provider "aws" {
  region = "us-east-1"
}

resource "aws_s3_bucket" "non_compliant_bucket" {
  bucket        = "aws-taggy-non-compliant"
  force_destroy = true

  tags = {
    # Prohibited tags
    "temp:test"     = "temporary-resource"
    "test:example"  = "should-not-be-here"
    "random:tag"    = "unexpected value"

    # Invalid tag keys (uppercase)
    "ENVIRONMENT"   = "PRODUCTION"
    "Owner"         = "john.doe@gmail.com"
    "Sensitive"     = "High Risk"

    # Incorrect required tags
    "project"       = "Invalid-Project-Name"
    "data_class"    = "unknown"
    "cost_center"   = "1234"

    # Additional problematic tags
    "RANDOM_TAG"    = "unexpected value"
  }
}

# Optional: Add some bucket configurations
resource "aws_s3_bucket_versioning" "versioning" {
  bucket = aws_s3_bucket.non_compliant_bucket.id
  versioning_configuration {
    status = "Enabled"
  }
}

# Outputs for the non-compliant S3 bucket
output "bucket_id" {
  description = "The ID of the non-compliant S3 bucket"
  value       = aws_s3_bucket.non_compliant_bucket.id
}

output "bucket_arn" {
  description = "The ARN of the non-compliant S3 bucket"
  value       = aws_s3_bucket.non_compliant_bucket.arn
}

output "bucket_region" {
  description = "The region of the non-compliant S3 bucket"
  value       = aws_s3_bucket.non_compliant_bucket.region
}

output "bucket_name" {
  description = "The name of the non-compliant S3 bucket"
  value       = aws_s3_bucket.non_compliant_bucket.bucket
}
