package resource

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/pkg/errors"
)

type AWSClient struct {
	EC2conn *ec2.EC2
	ASconn  *autoscaling.AutoScaling
	ELBconn *elb.ELB
	R53conn *route53.Route53
	CFconn  *cloudformation.CloudFormation
	EFSconn *efs.EFS
	IAMconn *iam.IAM
	KMSconn *kms.KMS
	S3conn  *s3.S3
	STSconn *sts.STS
}

// ApiDesc stores the necessary information about
// resource types (identified by its terraform type)
// to list and delete its resources via the go-aws-sdk
// and Terraform AWS provider API.
type ApiDesc struct {
	TerraformType      string
	DescribeOutputName []string
	DeleteId           string
	DescribeFn         interface{}
	DescribeFnInput    interface{}
	Select             func(Resources, interface{}, Filter, *AWSClient) []Resources
}

/*type Resources struct {
	Type  string // we use the terraform type for identification
	Ids   []*string
	Attrs []*map[string]string
	Tags  []*map[string]string
	Raw   interface{}
}*/

type Resources []*Resource

type Resource struct {
	Type  string // we use the terraform type for identification
	Id    string
	Attrs map[string]string
	Tags  map[string]string
}

// Supported returns for all supported
// resource types the API information
// to list (go-sdk API) and delete (AWS Terraform provider API)
// corresponding resources.
func Supported(c *AWSClient) []ApiDesc {
	return []ApiDesc{
		{
			"aws_autoscaling_group",
			[]string{"AutoScalingGroups"},
			"AutoScalingGroupName",
			c.ASconn.DescribeAutoScalingGroups,
			&autoscaling.DescribeAutoScalingGroupsInput{},
			filterGeneric,
		},
		{
			"aws_launch_configuration",
			[]string{"LaunchConfigurations"},
			"LaunchConfigurationName",
			c.ASconn.DescribeLaunchConfigurations,
			&autoscaling.DescribeLaunchConfigurationsInput{},
			filterGeneric,
		},
		{
			"aws_instance",
			[]string{"Reservations", "Instances"},
			"InstanceId",
			c.EC2conn.DescribeInstances,
			&ec2.DescribeInstancesInput{
				Filters: []*ec2.Filter{
					{
						Name: aws.String("instance-state-name"),
						Values: []*string{
							aws.String("pending"), aws.String("running"),
							aws.String("stopping"), aws.String("stopped"),
						},
					},
				},
			},
			filterGeneric,
		},
		{
			"aws_key_pair",
			[]string{"KeyPairs"},
			"KeyName",
			c.EC2conn.DescribeKeyPairs,
			&ec2.DescribeKeyPairsInput{},
			filterGeneric,
		},
		{
			"aws_elb",
			[]string{"LoadBalancerDescriptions"},
			"LoadBalancerName",
			c.ELBconn.DescribeLoadBalancers,
			&elb.DescribeLoadBalancersInput{},
			filterGeneric,
		},
		{
			"aws_vpc_endpoint",
			[]string{"VpcEndpoints"},
			"VpcEndpointId",
			c.EC2conn.DescribeVpcEndpoints,
			&ec2.DescribeVpcEndpointsInput{},
			filterGeneric,
		},
		// support tags
		{
			"aws_nat_gateway",
			[]string{"NatGateways"},
			"NatGatewayId",
			c.EC2conn.DescribeNatGateways,
			&ec2.DescribeNatGatewaysInput{},
			filterNatGateways,
		},
		{
			"aws_cloudformation_stack",
			[]string{"Stacks"},
			"StackId",
			c.CFconn.DescribeStacks,
			&cloudformation.DescribeStacksInput{},
			filterGeneric,
		},
		{
			"aws_route53_zone",
			[]string{"HostedZones"},
			"Id",
			c.R53conn.ListHostedZones,
			&route53.ListHostedZonesInput{},
			filterRoute53Zone,
		},
		{
			"aws_efs_file_system",
			[]string{"FileSystems"},
			"FileSystemId",
			c.EFSconn.DescribeFileSystems,
			&efs.DescribeFileSystemsInput{},
			filterEfsFileSystem,
		},
		//{
		//	"aws_route53_record",
		//	"ResourceRecordSets",
		//	"bla",
		//	c.r53conn.ListResourceRecordSets,
		//	&route53.ListResourceRecordSetsInput{},
		//	filterRoute53Record,
		//},
		// Elastic network interface (ENI) resource
		// sort by owner of the network interface?
		// support tags
		// attached to subnet
		{
			"aws_network_interface",
			[]string{"NetworkInterfaces"},
			"NetworkInterfaceId",
			c.EC2conn.DescribeNetworkInterfaces,
			&ec2.DescribeNetworkInterfacesInput{},
			filterGeneric,
		},
		{
			"aws_eip",
			[]string{"Addresses"},
			"AllocationId",
			c.EC2conn.DescribeAddresses,
			&ec2.DescribeAddressesInput{},
			filterGeneric,
		},
		{
			"aws_internet_gateway",
			[]string{"InternetGateways"},
			"InternetGatewayId",
			c.EC2conn.DescribeInternetGateways,
			&ec2.DescribeInternetGatewaysInput{},
			filterInternetGateways,
		},
		{
			"aws_subnet",
			[]string{"Subnets"},
			"SubnetId",
			c.EC2conn.DescribeSubnets,
			&ec2.DescribeSubnetsInput{},
			filterGeneric,
		},
		{
			"aws_route_table",
			[]string{"RouteTables"},
			"RouteTableId",
			c.EC2conn.DescribeRouteTables,
			&ec2.DescribeRouteTablesInput{},
			filterGeneric,
		},
		{
			"aws_security_group",
			[]string{"SecurityGroups"},
			"GroupId",
			c.EC2conn.DescribeSecurityGroups,
			&ec2.DescribeSecurityGroupsInput{},
			filterGeneric,
		},
		{
			"aws_network_acl",
			[]string{"NetworkAcls"},
			"NetworkAclId",
			c.EC2conn.DescribeNetworkAcls,
			&ec2.DescribeNetworkAclsInput{},
			filterGeneric,
		},
		{
			"aws_vpc",
			[]string{"Vpcs"},
			"VpcId",
			c.EC2conn.DescribeVpcs,
			&ec2.DescribeVpcsInput{},
			filterGeneric,
		},
		{
			"aws_iam_policy",
			[]string{"Policies"},
			"Arn",
			c.IAMconn.ListPolicies,
			&iam.ListPoliciesInput{},
			filterIamPolicy,
		},
		{
			"aws_iam_group",
			[]string{"Groups"},
			"GroupName",
			c.IAMconn.ListGroups,
			&iam.ListGroupsInput{},
			filterGeneric,
		},
		{
			"aws_iam_user",
			[]string{"Users"},
			"UserName",
			c.IAMconn.ListUsers,
			&iam.ListUsersInput{},
			filterIamUser,
		},
		{
			"aws_iam_role",
			[]string{"Roles"},
			"RoleName",
			c.IAMconn.ListRoles,
			&iam.ListRolesInput{},
			filterIamRole,
		},
		{
			"aws_iam_instance_profile",
			[]string{"InstanceProfiles"},
			"InstanceProfileName",
			c.IAMconn.ListInstanceProfiles,
			&iam.ListInstanceProfilesInput{},
			filterInstanceProfiles,
		},
		{
			"aws_kms_alias",
			[]string{"Aliases"},
			"AliasName",
			c.KMSconn.ListAliases,
			&kms.ListAliasesInput{},
			filterGeneric,
		},
		{
			"aws_kms_key",
			[]string{"Keys"},
			"KeyId",
			c.KMSconn.ListKeys,
			&kms.ListKeysInput{},
			filterKmsKeys,
		},
		{
			"aws_s3_bucket",
			[]string{"Buckets"},
			"Name",
			c.S3conn.ListBuckets,
			&s3.ListBucketsInput{},
			filterGeneric,
		},
		{
			"aws_ebs_snapshot",
			[]string{"Snapshots"},
			"SnapshotId",
			c.EC2conn.DescribeSnapshots,
			&ec2.DescribeSnapshotsInput{},
			filterSnapshots,
		},
		{
			"aws_ebs_volume",
			[]string{"Volumes"},
			"VolumeId",
			c.EC2conn.DescribeVolumes,
			&ec2.DescribeVolumesInput{},
			filterGeneric,
		},
		{
			"aws_ami",
			[]string{"Images"},
			"ImageId",
			c.EC2conn.DescribeImages,
			&ec2.DescribeImagesInput{},
			filterAmis,
		},
	}
}

// getSupported returns the apiInfo by the name of
// a given resource type
func getSupported(resType string, c *AWSClient) (ApiDesc, error) {
	for _, apiInfo := range Supported(c) {
		if apiInfo.TerraformType == resType {
			return apiInfo, nil
		}
	}
	return ApiDesc{}, errors.Errorf("no ApiDesc found for resource type %s", resType)
}
