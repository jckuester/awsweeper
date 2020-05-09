resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags {
    foo = "bar"
    Name = "awsweeper-testacc"
  }
}

resource "aws_subnet" "bar" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "10.1.1.0/24"

  tags {
    foo = "bar"
    Name = "awsweeper-testacc"
  }
}