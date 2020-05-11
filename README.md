<p align="center">
  <img alt="AWSweeper Logo" src="https://github.com/cloudetc/awsweeper/blob/master/img/logo.png" height="180" />
  <h3 align="center">AWSweeper</h3>
  <p align="center">A tool for cleaning your AWS account</p>
</p>

---
[![Release](https://img.shields.io/github/release/cloudetc/awsweeper.svg?style=for-the-badge)](https://github.com/cloudetc/awsweeper/releases/latest)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](/LICENSE.md)
[![Travis](https://img.shields.io/travis/cloudetc/awsweeper/master.svg?style=for-the-badge)](https://travis-ci.org/cloudetc/awsweeper)

AWSweeper cleans out all (or parts) of the resources in your AWS account. Resources to be deleted can be filtered by
their ID, tags or creation date using [regular expressions](https://golang.org/pkg/regexp/syntax/) declared via a filter
in a YAML file (see [filter.yml](example/config.yml) as an example).

AWSweeper [can delete many](#supported-resources), but not all resources yet. Your help
to support more resources is very much appreciated ([please read this issue](https://github.com/cloudetc/awsweeper/issues/21)
 to see how easy it is). 

Happy erasing!

[![AWSweeper tutorial](img/asciinema-tutorial.gif)](https://asciinema.org/a/149097)

## Installation

It's recommended to install a specific version of AWSweeper available on the
[releases page](https://github.com/cloudetc/awsweeper/releases).

Here is the recommended way to install AWSweeper v0.8.0:

```bash
# install it into ./bin/
curl -sSfL https://raw.githubusercontent.com/cloudetc/awsweeper/master/install.sh | sh -s v0.8.0
```

## Usage

    awsweeper [options] <filter.yml>

To see options available run `awsweeper --help`.

## Filter

Resources are deleted via a filter declared in a YAML file.

    aws_instance:
      # instance filter part 1
      - id: ^foo.*
        created:
          before: 2018-10-14
          after: 2018-06-28 12:28:39
            
      # instance filter part 2   
      - tags:
          foo: bar
          NOT(owner): .*
           
    aws_security_groups:

The filter snippet above deletes all EC2 instances that ID matches `^foo.*` and that have been created between
 `2018-06-28 12:28:39` and `2018-10-14` UTC (instance filter part 1); additionally, EC2 instances having a tag 
 `foo: bar` *AND* not a tag key `owner` with any value are deleted (instance filter part 2); last but not least,
 ALL security groups are deleted by this filter.

The general filter syntax is as follows:

    <resource type>:
      - id: <regex to filter by id> | NOT(<regex to filter by id>)
        tagged: bool (optional)
        tags:
          <key> | NOT(key): <regex to filter value> | NOT(<regex to filter value>)
          ...
        created:
          before: <timestamp> (optional)
          after: <timestamp> (optional)
      # OR
      - ...
    <resource type>:
      ...

Here is a more detailed description of the various ways to filter resources:

##### 1) Delete all resources of a particular type

   [Terraform resource type indentifiers](https://www.terraform.io/docs/providers/aws/index.html) are used to delete 
   resources by type. The following filter snippet deletes *ALL* security groups, IAM roles, and EC2 instances:
   
    aws_security_group:
    aws_iam_role:
    aws_instance:
   
   Don't forget the `:` at the end of each line. Use the [all.yml](./all.yml), to delete all (currently supported)
   resources.

##### 2) Delete by tags

   If most of your resources have tags, this is probably the best way to filter them
   for deletion. **Be aware**: Not all resources [support tags](#supported-resources) yet and can be filtered this way.
      
   The key and the value part of the tag filter can be negated by a surrounding `NOT(...)`. This allows for removing of 
   all resources not matching some tag key or value. In the example below, all EC2 instances without the `owner: me`
   tag are deleted:

    aws_instance:
      - tags:
          NOT(Owner): me
          
   The flag `tagged: false` deletes all resources that have no tags. Contrary, resources with any tags can be deleted 
   with `tagged: true`:

    aws_instance:
      - tagged: true

##### 3) Delete By ID

   You can narrow down on particular types of resources by filtering on based their IDs.

   To see what the ID of a resource is (could be its name, ARN, a random number),
   run AWSweeper in dry-run mode: `awsweeper --dry-run all.yml`. This way, nothing is deleted but
   all the IDs and tags of your resources are printed. Then, use this information to create the YAML config file.

   The id filter can be negated by surrounding the regex with `NOT(...)`

##### 4) By creation date

   You can select resources by filtering on the date they have been created using an absolute or relative date.

   The supported formats are:
   * Relative
     * Nanosecond: `1ns`
     * Microsecond: `1us`
     * Millisecond: `1ms`
     * Second: `1s`
     * Minute: `1m`
     * Hour: `1h`
     * Day: `1d`
     * Week: `1w`
     * Month: `1M`
     * Year: `1y`
   * Absolute:
     * RCF3339Nano, short dates: `2006-1-2T15:4:5.999999999Z07:00`
     * RFC3339Nano, short date, lower-case "t": `2006-1-2t15:4:5.999999999Z07:00`
     * Space separated, no time zone: `2006-1-2 15:4:5.999999999`
     * Date only: `2006-1-2`

## Dry-run mode

 Use `awsweeper --dry-run <filter.yml>` to only show what
would be deleted. This way, you can fine-tune your YAML filter configuration until it works the way you want it to.

## Supported resources

AWSweeper can currently delete more than 30 AWS resource types.

Note that the resource types in the list below are [Terraform Types](https://www.terraform.io/docs/providers/aws/index.html),
which must be used in the YAML configuration to filter resources.
A technical reason for this is that AWSweeper is build upon the already existing delete routines provided by the [Terraform AWS provider](https://github.com/terraform-providers/terraform-provider-aws).

| Resource Type                    | Delete by tag | Delete by creation date
| :-----------------------------   |:-------------:|:-----------------------:
| aws_ami                          | x             | x
| aws_autoscaling_group            | x             | x
| aws_cloudformation_stack         | x             | x
| aws_cloudtrail             |               |
| aws_cloudwatch_log_group (*new*) |               | x
| aws_ebs_snapshot                 | x             | x
| aws_ebs_volume                   | x             | x
| aws_ecs_cluster (*new*)          | x             |
| aws_efs_file_system              | x             | x
| aws_eip                          | x             |
| aws_elb                          | x             | x
| aws_iam_group                    | x             | x
| aws_iam_instance_profile         |               | x
| aws_iam_policy                   |               | x
| aws_iam_role                     | x             | x
| aws_iam_user                     | x             | x
| aws_instance                     | x             | x
| aws_internet_gateway             | x             |
| aws_key_pair                     | x             |
| aws_kms_alias                    |               |
| aws_kms_key                      |               |
| aws_lambda_function (*new*)      |               |
| aws_launch_configuration         |               | x
| aws_nat_gateway                  | x             |
| aws_network_acl                  | x             |
| aws_network_interface            | x             |
| aws_rds_instance (*new*)         |               | x
| aws_route53_zone                 |               |
| aws_route_table                  | x             |
| aws_s3_bucket                    |               | x
| aws_security_group               | x             |
| aws_subnet                       | x             |
| aws_vpc                          | x             |
| aws_vpc_endpoint                 | x             | x
   
## Acceptance tests

***IMPORTANT:*** Acceptance tests create real resources that might cost you money. Also, note that if you contribute a PR, 
the [Travis build](https://travis-ci.org/github/cloudetc/awsweeper) will always fail since AWS credentials are not 
injected into the PR build coming from forks for security reasons. You can either run tests locally against your personal 
AWS account or ask me to run them for you instead.

Run all acceptance tests with

    AWS_PROFILE=<myaccount> AWS_DEFAULT_REGION=us-west-2 make test-all

or to test the working of AWSweeper for a just single resource, such as `aws_vpc`, use

    AWS_PROFILE=<myaccount> AWS_DEFAULT_REGION=us-west-2 make test-all TESTARGS='-run=TestAcc_Vpc*'

## Disclaimer

This tool is thoroughly tested. However, you are using this tool at your own risk!
I will not take responsibility if you delete any critical resources in your
production environments.
