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

locals {
  az = "${var.region}c"
}

resource "aws_autoscaling_group" "test" {
  availability_zones = [local.az]
  desired_capacity   = 1
  max_size           = 1
  min_size           = 1

  launch_template {
    id      = aws_launch_template.test.id
    version = "$Latest"
  }
}

resource "aws_autoscaling_group" "test_tag" {
  availability_zones = [local.az]
  desired_capacity   = 1
  max_size           = 1
  min_size           = 1

  launch_template {
    id      = aws_launch_template.test.id
    version = "$Latest"
  }

  tag {
    key                 = "awsweeper"
    value               = "test-acc"
    propagate_at_launch = true
  }
}

resource "aws_autoscaling_group" "test_tags" {
  availability_zones = [local.az]
  desired_capacity   = 1
  max_size           = 1
  min_size           = 1

  launch_template {
    id      = aws_launch_template.test.id
    version = "$Latest"
  }

  tags =  [{
    key                 = "awsweeper"
    value               = "test-acc"
    propagate_at_launch = true
  }]
}

resource "aws_launch_template" "test" {
  name_prefix   = "foobar"
  image_id      = data.aws_ami.amazon_linux_2.image_id
  instance_type = "t2.micro"
}

data "aws_ami" "amazon_linux_2" {
  most_recent = true

  owners = ["amazon"]

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm*"]
  }
}
