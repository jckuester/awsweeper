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

resource "aws_iam_user" "test" {
  name        = var.name

  tags = {
    awsweeper = "test-acc"
  }
}

resource "aws_iam_role" "test" {
  name        = var.name

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
  name        = var.name
}

resource "aws_iam_policy" "test" {
  name        = var.name
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
  name        = var.name
  users      = [aws_iam_user.test.name]
  roles      = [aws_iam_role.test.name]
  groups     = [aws_iam_group.test.name]
  policy_arn = aws_iam_policy.test.arn
}
