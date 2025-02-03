provider "aws" {
  region = "us-east-1"
}

resource "aws_s3_bucket" "compliance_test_bucket" {
  bucket        = "aws-taggy"
  force_destroy = true

  tags = {
    Name                = "aws-taggy"
    Environment         = "development"
    Owner               = "data-engineering-team@company.com"
    Project             = "aws-taggy-demo"
    DataClassification  = "internal"
    CostCenter          = "DE-1234"
  }
}

# Optional: Add some bucket configurations
resource "aws_s3_bucket_versioning" "versioning" {
  bucket = aws_s3_bucket.compliance_test_bucket.id
  versioning_configuration {
    status = "Enabled"
  }
}
