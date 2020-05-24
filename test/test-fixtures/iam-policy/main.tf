provider "aws" {
  version = "~> 2.0"

  profile = var.profile
  region  = var.region
}

terraform {
  # The configuration for this backend will be filled in by Terragrunt
  backend "s3" {
  }
}

resource "aws_iam_user" "test" {
  name       = "awsweeper-test-acc"

  tags = {
    awsweeper = "test-acc"
  }
}

resource "aws_iam_role" "test" {
  name       = "awsweeper-test-acc"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

  tags = {
    awsweeper = "test-acc"
  }
}

resource "aws_iam_group" "test" {
  name       = "awsweeper-test-acc"
}

resource "aws_iam_policy" "test" {
  name       = "awsweeper-test-acc"
  description = "A test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy_attachment" "test" {
  name       = "awsweeper-test-acc"
  users      = [aws_iam_user.test.name]
  roles      = [aws_iam_role.test.name]
  groups     = [aws_iam_group.test.name]
  policy_arn = aws_iam_policy.test.arn
}