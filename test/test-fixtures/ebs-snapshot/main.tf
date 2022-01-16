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

resource "aws_ebs_volume" "test" {
  availability_zone = "${var.region}a"
  size              = 1

  tags = {
    awsweeper = "test-acc"
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id

  tags = {
    awsweeper = "test-acc"
  }
}
