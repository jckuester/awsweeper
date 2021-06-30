provider "aws" {
  version = "~> 3.0"

  profile = var.profile
  region  = var.region
}

terraform {
  # The configuration for this backend will be filled in by Terragrunt
  backend "s3" {
  }
}

resource "aws_iam_role" "test" {
  name = var.name

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_lambda_function" "test" {
  filename      = "lambda_function_payload.zip"
  function_name = var.name
  role          = aws_iam_role.test.arn
  handler       = "exports.test"

 source_code_hash = filebase64sha256("lambda_function_payload.zip")

  runtime = "python3.7"

  environment {
    variables = {
      awsweeper = "test-acc"
    }
  }

  tags = {
    awsweeper = "test-acc"
  }
}
