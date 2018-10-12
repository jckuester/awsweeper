package resource

import (
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/efs/efsiface"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/go-errors/errors"
)

// TerraformResourceType identifies a type of resource
// TerraformResourceType identifies a type of resource
type TerraformResourceType string

const (
	AutoscalingGroup    TerraformResourceType = "aws_autoscaling_group"
	LaunchConfiguration TerraformResourceType = "aws_launch_configuration"
	Instance            TerraformResourceType = "aws_instance"
	KeyPair             TerraformResourceType = "aws_key_pair"
	Elb                 TerraformResourceType = "aws_elb"
	VpcEndpoint         TerraformResourceType = "aws_vpc_endpoint"
	NatGateway          TerraformResourceType = "aws_nat_gateway"
	CloudformationStack TerraformResourceType = "aws_cloudformation_stack"
	Route53Zone         TerraformResourceType = "aws_route53_zone"
	EfsFileSystem       TerraformResourceType = "aws_efs_file_system"
	NetworkInterface    TerraformResourceType = "aws_network_interface"
	Eip                 TerraformResourceType = "aws_eip"
	InternetGateway     TerraformResourceType = "aws_internet_gateway"
	Subnet              TerraformResourceType = "aws_subnet"
	RouteTable          TerraformResourceType = "aws_route_table"
	SecurityGroup       TerraformResourceType = "aws_security_group"
	NetworkAcl          TerraformResourceType = "aws_network_acl"
	Vpc                 TerraformResourceType = "aws_vpc"
	IamPolicy           TerraformResourceType = "aws_iam_policy"
	IamGroup            TerraformResourceType = "aws_iam_group"
	IamUser             TerraformResourceType = "aws_iam_user"
	IamRole             TerraformResourceType = "aws_iam_role"
	IamInstanceProfile  TerraformResourceType = "aws_iam_instance_profile"
	KmsAlias            TerraformResourceType = "aws_kms_alias"
	KmsKey              TerraformResourceType = "aws_kms_key"
	S3Bucket            TerraformResourceType = "aws_s3_bucket"
	EbsSnapshot         TerraformResourceType = "aws_ebs_snapshot"
	EbsVolume           TerraformResourceType = "aws_ebs_volume"
	Ami                 TerraformResourceType = "aws_ami"
)

var deleteIDs = map[TerraformResourceType]string{
	AutoscalingGroup:    "AutoScalingGroupName",
	LaunchConfiguration: "LaunchConfigurationName",
}

func getDeleteID(resType TerraformResourceType) (string, error) {
	deleteID, found := deleteIDs[resType]
	if !found {
		return "", errors.Errorf("no delete ID specified for resource type: %s", resType)
	}
	return deleteID, nil
}

// AWS wraps the AWS API
type AWS struct {
	ec2iface.EC2API
	autoscalingiface.AutoScalingAPI
	elbiface.ELBAPI
	route53iface.Route53API
	cloudformationiface.CloudFormationAPI
	efsiface.EFSAPI
	iamiface.IAMAPI
	kmsiface.KMSAPI
	s3iface.S3API
	stsiface.STSAPI
}

// NewAWS creates an AWS instance
func NewAWS(s *session.Session) *AWS {
	return &AWS{
		AutoScalingAPI:    autoscaling.New(s),
		CloudFormationAPI: cloudformation.New(s),
		EC2API:            ec2.New(s),
		EFSAPI:            efs.New(s),
		ELBAPI:            elb.New(s),
		IAMAPI:            iam.New(s),
		KMSAPI:            kms.New(s),
		Route53API:        route53.New(s),
		S3API:             s3.New(s),
		STSAPI:            sts.New(s),
	}
}

// rawResources is a list of AWS resources.
type Resources []*Resource

// Resource contains information about
// a single AWS resource.
type Resource struct {
	Type  TerraformResourceType
	ID    string
	Attrs map[string]string
	Tags  map[string]string
}

// rawResources list all resources of a particular type
func (aws *AWS) rawResources(resType TerraformResourceType) (interface{}, error) {
	switch resType {
	case AutoscalingGroup:
		return aws.autoscalingGroups()
	case LaunchConfiguration:
		return aws.launchConfigurations()
	default:
		return nil, errors.Errorf("unknown or unsupported resource type: %s", resType)
	}
}

//	return []APIDesc{
//		{
//			AutoscalingGroup,
//			"AutoScalingGroupName",
//			defaultFilter,
//		},
//		{
//			LaunchConfiguration,
//			"LaunchConfigurationName",
//			defaultFilter,
//		},
//		{
//			Instance,
//			[]string{"Reservations", "Instances"},
//			"InstanceId",
//			c.EC2conn.DescribeInstances,
//			&ec2.DescribeInstancesInput{
//				Filters: []*ec2.Filter{
//					{
//						Name: aws.String("instance-state-name"),
//						Values: []*string{
//							aws.String("pending"), aws.String("running"),
//							aws.String("stopping"), aws.String("stopped"),
//						},
//					},
//				},
//			},
//			defaultFilter,
//		},
//		{
//			KeyPair,
//			[]string{"KeyPairs"},
//			"KeyName",
//			c.EC2conn.DescribeKeyPairs,
//			&ec2.DescribeKeyPairsInput{},
//			defaultFilter,
//		},
//		{
//			Elb,
//			[]string{"LoadBalancerDescriptions"},
//			"LoadBalancerName",
//			c.ELBconn.DescribeLoadBalancers,
//			&elb.DescribeLoadBalancersInput{},
//			defaultFilter,
//		},
//		{
//			VpcEndpoint,
//			[]string{"VpcEndpoints"},
//			"VpcEndpointId",
//			c.EC2conn.DescribeVpcEndpoints,
//			&ec2.DescribeVpcEndpointsInput{},
//			defaultFilter,
//		},
//		{
//			// TODO support findTags
//			NatGateway,
//			[]string{"NatGateways"},
//			"NatGatewayId",
//			c.EC2conn.DescribeNatGateways,
//			&ec2.DescribeNatGatewaysInput{
//				Filter: []*ec2.Filter{
//					{
//						Name: aws.String("state"),
//						Values: []*string{
//							aws.String("available"),
//						},
//					},
//				},
//			},
//			defaultFilter,
//		},
//		{
//			CloudformationStack,
//			[]string{"Stacks"},
//			"StackId",
//			c.CFconn.DescribeStacks,
//			&cloudformation.DescribeStacksInput{},
//			defaultFilter,
//		},
//		{
//			Route53Zone,
//			[]string{"HostedZones"},
//			"Id",
//			c.R53conn.ListHostedZones,
//			&route53.ListHostedZonesInput{},
//			defaultFilter,
//		},
//		{
//			EfsFileSystem,
//			[]string{"FileSystems"},
//			"FileSystemId",
//			c.EFSconn.DescribeFileSystems,
//			&efs.DescribeFileSystemsInput{},
//			efsFileSystemFilter,
//		},
//		// Elastic network interface (ENI) resource
//		// sort by owner of the network interface?
//		// support findTags
//		// attached to subnet
//		{
//			NetworkInterface,
//			[]string{"NetworkInterfaces"},
//			"NetworkInterfaceId",
//			c.EC2conn.DescribeNetworkInterfaces,
//			&ec2.DescribeNetworkInterfacesInput{},
//			defaultFilter,
//		},
//		{
//			Eip,
//			[]string{"Addresses"},
//			"AllocationId",
//			c.EC2conn.DescribeAddresses,
//			&ec2.DescribeAddressesInput{},
//			defaultFilter,
//		},
//		{
//			InternetGateway,
//			[]string{"InternetGateways"},
//			"InternetGatewayId",
//			c.EC2conn.DescribeInternetGateways,
//			&ec2.DescribeInternetGatewaysInput{},
//			defaultFilter,
//		},
//		{
//			Subnet,
//			[]string{"Subnets"},
//			"SubnetId",
//			c.EC2conn.DescribeSubnets,
//			&ec2.DescribeSubnetsInput{},
//			defaultFilter,
//		},
//		{
//			RouteTable,
//			[]string{"RouteTables"},
//			"RouteTableId",
//			c.EC2conn.DescribeRouteTables,
//			&ec2.DescribeRouteTablesInput{},
//			defaultFilter,
//		},
//		{
//			SecurityGroup,
//			[]string{"SecurityGroups"},
//			"GroupId",
//			c.EC2conn.DescribeSecurityGroups,
//			&ec2.DescribeSecurityGroupsInput{},
//			defaultFilter,
//		},
//		{
//			NetworkAcl,
//			[]string{"NetworkAcls"},
//			"NetworkAclId",
//			c.EC2conn.DescribeNetworkAcls,
//			&ec2.DescribeNetworkAclsInput{},
//			defaultFilter,
//		},
//		{
//			Vpc,
//			[]string{"Vpcs"},
//			"VpcId",
//			c.EC2conn.DescribeVpcs,
//			&ec2.DescribeVpcsInput{},
//			defaultFilter,
//		},
//		{
//			IamPolicy,
//			[]string{"Policies"},
//			"Arn",
//			c.IAMconn.ListPolicies,
//			&iam.ListPoliciesInput{},
//			iamPolicyFilter,
//		},
//		{
//			IamGroup,
//			[]string{"Groups"},
//			"GroupName",
//			c.IAMconn.ListGroups,
//			&iam.ListGroupsInput{},
//			defaultFilter,
//		},
//		{
//			IamUser,
//			[]string{"Users"},
//			"UserName",
//			c.IAMconn.ListUsers,
//			&iam.ListUsersInput{},
//			iamUserFilter,
//		},
//		{
//			IamRole,
//			[]string{"Roles"},
//			"RoleName",
//			c.IAMconn.ListRoles,
//			&iam.ListRolesInput{},
//			defaultFilter,
//		},
//		{
//			IamInstanceProfile,
//			[]string{"InstanceProfiles"},
//			"InstanceProfileName",
//			c.IAMconn.ListInstanceProfiles,
//			&iam.ListInstanceProfilesInput{},
//			defaultFilter,
//		},
//		{
//			KmsAlias,
//			[]string{"Aliases"},
//			"AliasName",
//			c.KMSconn.ListAliases,
//			&kms.ListAliasesInput{},
//			defaultFilter,
//		},
//		{
//			KmsKey,
//			[]string{"Keys"},
//			"KeyId",
//			c.KMSconn.ListKeys,
//			&kms.ListKeysInput{},
//			kmsKeysFilter,
//		},
//		{
//			S3Bucket,
//			[]string{"Buckets"},
//			"Name",
//			c.S3conn.ListBuckets,
//			&s3.ListBucketsInput{},
//			defaultFilter,
//		},
//		{
//			EbsSnapshot,
//			[]string{"Snapshots"},
//			"SnapshotId",
//			c.EC2conn.DescribeSnapshots,
//			&ec2.DescribeSnapshotsInput{
//				Filters: []*ec2.Filter{
//					{
//						Name: aws.String("owner-id"),
//						Values: []*string{
//							accountID(c),
//						},
//					},
//				},
//			},
//			defaultFilter,
//		},
//		{
//			EbsVolume,
//			[]string{"Volumes"},
//			"VolumeId",
//			c.EC2conn.DescribeVolumes,
//			&ec2.DescribeVolumesInput{},
//			defaultFilter,
//		},
//		{
//			Ami,
//			[]string{"Images"},
//			"ImageId",
//			c.EC2conn.DescribeImages,
//			&ec2.DescribeImagesInput{
//				Filters: []*ec2.Filter{
//					{
//						Name: aws.String("owner-id"),
//						Values: []*string{
//							accountID(c),
//						},
//					},
//				},
//			},
//			defaultFilter,
//		},
//	}
//}

func (aws *AWS) autoscalingGroups() (interface{}, error) {
	output, err := aws.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	if err != nil {
		return nil, err
	}
	return output.AutoScalingGroups, nil
}

func (aws *AWS) launchConfigurations() (interface{}, error) {
	output, err := aws.DescribeLaunchConfigurations(&autoscaling.DescribeLaunchConfigurationsInput{})
	if err != nil {
		return nil, err
	}
	return output.LaunchConfigurations, nil
}

// accountID returns the account ID of the AWS account
// for the currently used credentials or AWS profile, resp.
func (c *AWS) accountID() *string {
	res, err := c.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		log.Fatal(err)
	}
	return res.Account
}
