provider "aws" {
  version = "~> 2.0"

  profile = var.profile
  region = var.region
}

terraform {
  # The configuration for this backend will be filled in by Terragrunt
  backend "s3" {
  }
}

data "aws_caller_identity" "current" {}

resource "aws_cloudtrail" "test" {
  name = var.name
  s3_bucket_name = aws_s3_bucket.test.id
  include_global_service_events = false

  tags = {
    awsweeper = "test-acc"
  }
}

resource "aws_s3_bucket" "test" {
  bucket = var.name
  force_destroy = true

  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AWSCloudTrailAclCheck",
            "Effect": "Allow",
            "Principal": {
              "Service": "cloudtrail.amazonaws.com"
            },
            "Action": "s3:GetBucketAcl",
            "Resource": "arn:aws:s3:::${var.name}"
        },
        {
            "Sid": "AWSCloudTrailWrite",
            "Effect": "Allow",
            "Principal": {
                "Service": "cloudtrail.amazonaws.com"
            },
            "Action": "s3:PutObject",
            "Resource": "arn:aws:s3:::${var.name}/AWSLogs/${data.aws_caller_identity.current.account_id}/*",
            "Condition": {
                "StringEquals": {
                    "s3:x-amz-acl": "bucket-owner-full-control"
                }
            }
        }
    ]
}
POLICY
}
