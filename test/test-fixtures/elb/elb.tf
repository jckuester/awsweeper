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

resource "aws_elb" "test" {
  name = var.name
  subnets = [
    aws_default_subnet.test.id
  ]

  listener {
    instance_port = 80
    instance_protocol = "tcp"
    lb_port = 80
    lb_protocol = "tcp"
  }

  tags = {
    awsweeper = "test-acc"
  }
}

resource "aws_default_subnet" "test" {
  availability_zone = "${var.region}c"
}