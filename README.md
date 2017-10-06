# AWSweeper

Wipe out resources in your AWS account.

## Usage

    awsweeper <config.yaml>
    
Have a look at [test.yaml](test.integration/test.yaml) for an example of the configuration file.
    
## Supported resources

Here is list of [all the various resources you can create within AWS](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html).
AWSweeper can delete many but not all of them:

- aws_autoscaling_group
- aws_launch_configuration
- aws_instance
- aws_elb
- aws_vpc_endpoint
- aws_nat_gateway
- aws_cloudformation_stack
- aws_route53_zone
- aws_eip
- aws_internet_gateway
- aws_efs_file_system
- aws_network_interface
- aws_subnet
- aws_route_table
- aws_network_acl
- aws_security_group
- aws_vpc
- aws_iam_user
- aws_iam_role
- aws_iam_policy
- aws_iam_instance_profile
- aws_kms_alias
- aws_kms_key

## Tests

For integration testing, resources of each type are created with terraform. Then we run awsweeper with a test
configuration to delete all resources:

     # create resources
     terraform apply
     
     # delete resources
     go run *.go test.integration/test.yaml
     
     # check if all resources have been wiped properly
     terraform destroy

## Disclaimer

You are using this tool at your own risk! I will not take any responsibility if you delete any critical resources in your
production environments.
