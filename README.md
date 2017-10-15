# AWSweeper

AWSweeper wipes out all (or parts) of the resources in your AWS account. What to delete is controlled
with a yaml file (see [test.yml](test.integration/test.yml), for example). Resources can be filtered by their tags or IDs
using [regular expressions](https://golang.org/pkg/regexp/syntax/).

Currently, AWSweeper [can delete many](#supported-resources), but not all resources. We are working on it!

## Usage

    awsweeper [options] <config.yml>

To see the options available run `awsweeper --help`.
    
## Filter resources

Resources to be deleted are filtered with a yaml configuration.
Have a look at the following example:

    aws_security_group:
    aws_instance:
      tags:
        foo: bar
        bla: blub
    aws_iam_role:
      ids:
      - ^foo.*            

There are three ways of filtering resources for deletion:

1) Delete *all* resources of a particular type

   [Terraform types](https://www.terraform.io/docs/providers/aws/index.html) are used to identify resources of a certain type
   (e.g., `aws_security_group` filters for resources that are security groups, `aws_iam_role` for roles,
   or `aws_instance` for all EC2 instances).

   In the example above, by simply adding `security_group:` (no further filters for IDs or tags),
   all security groups in your account would be deleted. Use the [all.yml](./all.yml), to delete all (currently supported) 
   resources in your account.

2) Filter by tags

   If most of your resources have tags, this is probably the way to filter them 
   for deletion. But be aware: not all resources support tags and can be filtered this way.
   
   In the example above, all instances are terminated that have either a tag with key `foo` and value `bar` or key `bla` and value `blub`, or both.
   
3) Filter by IDs
   
   To find out what the IDs of your resources are (sometimes their name, sometimes an ARN, or random number),
   run awsweeper in test-run mode: `awsweeper --test-run <config.yml>`. This way, nothing is deleted but
   all the IDs and tags or your resources will be printed. Then, use them to create the config.   

## Test run

 Use `awsweeper --test-run <config.yml>` to only show what
would be deleted. This way, you can iterate on the configuration until it works the way you want it to. 

## Supported resources

Here is list of [all the various types of resources you can create within AWS](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html).
AWSweeper can currently delete many but not all of the existing resource types:

- aws_autoscaling_group
- aws_cloudformation_stack
- aws_efs_file_system
- aws_eip
- aws_elb
- aws_iam_instance_profile
- aws_iam_policy
- aws_iam_role
- aws_iam_user
- aws_instance
- aws_internet_gateway
- aws_kms_alias
- aws_kms_key
- aws_launch_configuration
- aws_nat_gateway
- aws_network_acl
- aws_network_interface
- aws_route53_zone
- aws_route_table
- aws_security_group
- aws_subnet
- aws_vpc
- aws_vpc_endpoint

Note that the above list contains [terraform types](https://www.terraform.io/docs/providers/aws/index.html. They 
can be used as identifiers for resources tpyes in the yaml configuration. The reason is that AWSweeper 
is build upon delete functions provided by the [Terraform AWS provider](https://github.com/terraform-providers/terraform-provider-aws).

## Tests

Integration testing is semi-automated for now. Resources of each type are created with terraform. Then awsweeper is used with a test
configuration to delete all resources again:

     # create resources
     cd test.integration; terraform apply
     
     # delete resources
     go run ../*.go test.integration/test.yml
     
     # check if all resources have been wiped properly
     terraform destroy

## Disclaimer

You are using this tool at your own risk! I will not take any responsibility if you delete any critical resources in your
production environments. Use it for your test accounts only.
