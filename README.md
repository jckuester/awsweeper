[![AWSweeper tutorial](img/asciinema-tutorial.gif)](https://asciinema.org/a/149097)


# AWSweeper

<p align="right">
  <a href="https://goreportcard.com/report/github.com/cloudetc/awsweeper">
  <img src="https://goreportcard.com/badge/github.com/cloudetc/awsweeper" /></a>
</p>

AWSweeper wipes out all (or parts) of the resources in your AWS account. Resources to be deleted can be filtered by their tags or IDs
using [regular expressions](https://golang.org/pkg/regexp/syntax/) declared in a yaml file (see [config.yml](dependency/config.yml)).

AWSweeper [can delete many](#supported-resources), but not all resources yet.

We are working on it. Happy erasing!

## Download

Releases for your platform are [here](https://github.com/cloudetc/awsweeper/releases).

## Usage

    awsweeper [options] <config.yml>

To see options available run `awsweeper --help`.
    
## Filter resources for deletion

Resources to be deleted are filtered by a yaml configuration. To learn how, have a look at the following example:

    aws_security_group:
    aws_instance:
      tags:
        foo: bar
        bla: blub
    aws_iam_role:
      ids:
      - ^foo.*            

There are three ways to filter resources:

1) All resources of a particular type

   [Terraform types](https://www.terraform.io/docs/providers/aws/index.html) are used to identify resources of a particular type
   (e.g., `aws_security_group` selects all resources that are security groups, `aws_iam_role` all roles,
   or `aws_instance` all EC2 instances).

   In the example above, by simply adding `security_group:` (no further filters for IDs or tags),
   all security groups in your account would be deleted. Use the [all.yml](./all.yml), to delete all (currently supported) 
   resources.

2) By tags

   You can narrow down on particular types of resources by the tags they have.

   If most of your resources have tags, this is probably the best to filter them 
   for deletion. But be aware: not all resources support tags and can be filtered this way.
   
   In the example above, all EC2 instances are terminated that have either a tag with key `foo` and value `bar` or key `bla` and value `blub`, or both.
   
3) By IDs
   
   You can narrow down on particular types of resources by filtering on their IDs.

   To see what the IDs of your resources are (could be their name, ARN, a random number),
   run awsweeper in dry-run mode: `awsweeper --dry-run all.yml`. This way, nothing is deleted but
   all the IDs and tags of your resources are printed. Then, use this information to create the yaml file.
   
   In the example above, all roles which name starts with `foo` are deleted (the ID of roles is their name).
   
## Test run

 Use `awsweeper --dry-run <config.yml>` to only show what
would be deleted. This way, you can fine-tune your yaml configuration until it works the way you want it to. 

## Supported resources

AWSweeper can currently delete many but not [all of the existing types of AWS resources](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html):

- aws_ami
- aws_autoscaling_group
- aws_cloudformation_stack
- aws_ebs_snapshot
- aws_ebs_volume
- aws_efs_file_system
- aws_eip
- aws_elb
- aws_iam_group
- aws_iam_instance_profile
- aws_iam_policy
- aws_iam_role
- aws_iam_user
- aws_instance
- aws_internet_gateway
- aws_key_pair (***new***)
- aws_kms_alias
- aws_kms_key
- aws_launch_configuration
- aws_nat_gateway
- aws_network_acl
- aws_network_interface
- aws_route53_zone
- aws_route_table
- aws_s3_bucket
- aws_security_group
- aws_subnet
- aws_vpc
- aws_vpc_endpoint

Note that the above list contains [terraform types](https://www.terraform.io/docs/providers/aws/index.html) which must be used instead of [AWS resource types](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html) to identify resources in the yaml configuration.
The reason is that AWSweeper is build upon the already existing delete routines provided by the [Terraform AWS provider](https://github.com/terraform-providers/terraform-provider-aws).

## Acceptance tests

***WARNING:*** Running acceptance tests create real resources that might cost you money.

Run all acceptance tests with

    make testacc

or use

    make testacc TESTARGS='-run=TestAccVpc*'

to test the working of AWSweeper for a just single resource, such as `aws_vpc`.

## Disclaimer

You are using this tool at your own risk! I will not take any responsibility if you delete any critical resources in your
production environments. Use it for your test accounts only.
