# Used to find the order in which supported resources have to be deleted

variable "profile" {
  description = "Use a specific profile from your credential file"
}

variable "region" {
  default = "us-west-2"
}

provider "aws" {
  version = ">= 0.1.4"

  region = "${var.region}"
  profile = "${var.profile}"
}

terraform {
  required_version = ">= 0.10.0"
}

data "aws_ami" "foo" {
  most_recent = true
  owners = ["099720109477"]

  filter {
    name = "name"
    values = ["*ubuntu-trusty-14.04-amd64-server-*"]
  }

  filter {
    name = "state"
    values = ["available"]
  }

  filter {
    name = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name = "is-public"
    values = ["true"]
  }
}

# resource "aws_ami" "foo" {}

resource "aws_autoscaling_group" "foo" {
  name_prefix = "foo-"
  max_size = "1"
  min_size = "1"

  launch_configuration = "${aws_launch_configuration.foo.id}"
  vpc_zone_identifier = ["${aws_subnet.foo.id}"]

  load_balancers = ["${aws_elb.foo.id}"]

  tag {
    key = "Name"
    value = "foo"
    propagate_at_launch = false
  }
}

# resource "aws_cloudformation_stack" "foo" {}

#aws_ebs_snapshot

#aws_ebs_volume

# aws_efs_file_system

resource "aws_eip" "foo" {
  vpc = true
  //  instance = "${aws_instance.foo.id}"
}

resource "aws_elb" "foo" {
  name = "foo"
  subnets = [ "${aws_subnet.foo.id}" ]
  security_groups = [ "${aws_security_group.foo.id}" ]

  listener {
    instance_port = 80
    instance_protocol = "tcp"
    lb_port = 80
    lb_protocol = "tcp"
  }

  # It seems tags don't exist for ELBs
  tags {
    Name = "foo"
  }
}

# deleted together with user
resource "aws_iam_access_key" "foo" {
  user = "${aws_iam_user.foo.name}"
}

resource "aws_iam_group" "group" {
  name = "foo_group"
}

resource "aws_iam_instance_profile" "foo" {
  name  = "foo_profile"
  role = "${aws_iam_role.foo.name}"
}

resource "aws_iam_policy" "policy" {
  name        = "foo_policy"
  path        = "/"
  description = "My test policy"

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

resource "aws_iam_role" "foo" {
  name = "foo_role"
  path = "/"

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
}

resource "aws_iam_user" "foo" {
  name = "foo"
  path = "/system/"
}

# inline policy attached directly to user
resource "aws_iam_user_policy" "foo" {
  name = "foo"
  user = "${aws_iam_user.foo.name}"

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

resource "aws_iam_user_policy_attachment" "foo" {
  user       = "${aws_iam_user.foo.name}"
  policy_arn = "${aws_iam_policy.policy.arn}"
}

resource "aws_instance" "foo" {
  ami = "${data.aws_ami.foo.id}"
  instance_type = "t2.micro"
  security_groups = ["${aws_security_group.foo.id}"]
  subnet_id = "${aws_subnet.foo.id}"

  tags {
    Name = "foo"
  }
}

resource "aws_internet_gateway" "foo" {
  vpc_id = "${aws_vpc.foo.id}"

  tags {
    Name = "foo"
  }
}

#aws_kms_alias

#aws_kms_key

resource "aws_launch_configuration" "foo" {
  name_prefix = "foo-"
  image_id = "${data.aws_ami.foo.id}"
  instance_type = "t2.micro"
  associate_public_ip_address = true

  security_groups = [
    "${aws_security_group.foo.id}"
  ]

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_nat_gateway" "foo" {
  allocation_id = "${aws_eip.foo.id}"
  subnet_id = "${aws_subnet.foo.id}"

  tags {
    Name = "foo"
  }
}

resource "aws_network_acl" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  subnet_ids = [ "${aws_subnet.foo.id}" ]

  tags {
    Name = "foo"
  }
}

resource "aws_network_interface" "foo" {
  subnet_id       = "${aws_subnet.foo.id}"
  security_groups = ["${aws_security_group.foo.id}"]

  attachment {
    instance     = "${aws_instance.foo.id}"
    device_index = 1
  }

  # it seems that tags are not supported
  tags {
    Name = "foo"
  }
}

#aws_route53_zone

resource "aws_route_table" "foo" {
  vpc_id = "${aws_vpc.foo.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.foo.id}"
  }

  tags {
    Name = "foo"
  }
}

# aws_s3_bucket

resource "aws_security_group" "foo" {
  name = "foo"
  description = "Allow traffic on port 80"
  vpc_id = "${aws_vpc.foo.id}"

  ingress {
    from_port = 80
    to_port = 80
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  tags {
    Name = "foo"
  }
}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "${cidrsubnet(aws_vpc.foo.cidr_block, 8, count.index + 10)}"
  availability_zone = "${var.region}a"

  tags {
    Name = "foo"
  }
}

resource "aws_vpc" "foo" {
  cidr_block = "10.0.0.0/16"

  tags {
    Name = "foo"
  }
}

resource "aws_vpc_endpoint" "foo" {
  vpc_id       = "${aws_vpc.foo.id}"
  service_name = "com.amazonaws.us-west-2.s3"
}
