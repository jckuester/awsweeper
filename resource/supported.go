package resource

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/aws/aws-sdk-go/service/cloudtrail/cloudtrailiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/efs/efsiface"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/pkg/errors"
)

// TerraformResourceType identifies the type of a resource
type TerraformResourceType string

const (
	Ami                 TerraformResourceType = "aws_ami"
	AutoscalingGroup    TerraformResourceType = "aws_autoscaling_group"
	CloudformationStack TerraformResourceType = "aws_cloudformation_stack"
	CloudWatchLogGroup  TerraformResourceType = "aws_cloudwatch_log_group"
	EbsSnapshot         TerraformResourceType = "aws_ebs_snapshot"
	EbsVolume           TerraformResourceType = "aws_ebs_volume"
	EcsCluster          TerraformResourceType = "aws_ecs_cluster"
	EfsFileSystem       TerraformResourceType = "aws_efs_file_system"
	Eip                 TerraformResourceType = "aws_eip"
	Elb                 TerraformResourceType = "aws_elb"
	IamGroup            TerraformResourceType = "aws_iam_group"
	IamInstanceProfile  TerraformResourceType = "aws_iam_instance_profile"
	IamPolicy           TerraformResourceType = "aws_iam_policy"
	IamRole             TerraformResourceType = "aws_iam_role"
	IamUser             TerraformResourceType = "aws_iam_user"
	Instance            TerraformResourceType = "aws_instance"
	InternetGateway     TerraformResourceType = "aws_internet_gateway"
	KeyPair             TerraformResourceType = "aws_key_pair"
	KmsAlias            TerraformResourceType = "aws_kms_alias"
	KmsKey              TerraformResourceType = "aws_kms_key"
	LambdaFunction      TerraformResourceType = "aws_lambda_function"
	LaunchConfiguration TerraformResourceType = "aws_launch_configuration"
	NatGateway          TerraformResourceType = "aws_nat_gateway"
	NetworkACL          TerraformResourceType = "aws_network_acl"
	NetworkInterface    TerraformResourceType = "aws_network_interface"
	DBInstance          TerraformResourceType = "aws_db_instance"
	Route53Zone         TerraformResourceType = "aws_route53_zone"
	RouteTable          TerraformResourceType = "aws_route_table"
	S3Bucket            TerraformResourceType = "aws_s3_bucket"
	SecurityGroup       TerraformResourceType = "aws_security_group"
	Subnet              TerraformResourceType = "aws_subnet"
	Trail               TerraformResourceType = "aws_cloudtrail"
	Vpc                 TerraformResourceType = "aws_vpc"
	VpcEndpoint         TerraformResourceType = "aws_vpc_endpoint"
)

var (
	deleteIDs = map[TerraformResourceType]string{
		Ami:                 "ImageId",
		AutoscalingGroup:    "AutoScalingGroupName",
		CloudformationStack: "StackId",
		CloudWatchLogGroup:  "LogGroupName",
		EbsSnapshot:         "SnapshotId",
		EbsVolume:           "VolumeId",
		// Note: to import a cluster, the name is used as ID
		EcsCluster:          "ClusterArn",
		EfsFileSystem:       "FileSystemId",
		Eip:                 "AllocationId",
		Elb:                 "LoadBalancerName",
		IamGroup:            "GroupName",
		IamInstanceProfile:  "InstanceProfileName",
		IamPolicy:           "Arn",
		IamRole:             "RoleName",
		IamUser:             "UserName",
		Instance:            "InstanceId",
		InternetGateway:     "InternetGatewayId",
		KeyPair:             "KeyName",
		KmsAlias:            "AliasName",
		KmsKey:              "KeyId",
		LambdaFunction:      "FunctionName",
		LaunchConfiguration: "LaunchConfigurationName",
		NatGateway:          "NatGatewayId",
		NetworkACL:          "NetworkAclId",
		NetworkInterface:    "NetworkInterfaceId",
		DBInstance:          "DBInstanceIdentifier",
		Route53Zone:         "Id",
		RouteTable:          "RouteTableId",
		S3Bucket:            "Name",
		SecurityGroup:       "GroupId",
		Subnet:              "SubnetId",
		Trail:               "Name",
		Vpc:                 "VpcId",
		VpcEndpoint:         "VpcEndpointId",
	}

	// DependencyOrder is the order in which resource types should be deleted,
	// since dependent resources need to be deleted before their dependencies
	// (e.g. aws_subnet before aws_vpc)
	DependencyOrder = map[TerraformResourceType]int{
		LambdaFunction:      10100,
		EcsCluster:          10000,
		AutoscalingGroup:    9990,
		Instance:            9980,
		KeyPair:             9970,
		Elb:                 9960,
		VpcEndpoint:         9950,
		NatGateway:          9940,
		CloudformationStack: 9930,
		Route53Zone:         9920,
		EfsFileSystem:       9910,
		LaunchConfiguration: 9900,
		Eip:                 9890,
		InternetGateway:     9880,
		Subnet:              9870,
		RouteTable:          9860,
		SecurityGroup:       9850,
		NetworkACL:          9840,
		Vpc:                 9830,
		DBInstance:          9825,
		IamPolicy:           9820,
		IamGroup:            9810,
		IamUser:             9800,
		IamRole:             9790,
		IamInstanceProfile:  9780,
		S3Bucket:            9750,
		Ami:                 9740,
		EbsVolume:           9730,
		EbsSnapshot:         9720,
		KmsAlias:            9610,
		KmsKey:              9600,
		NetworkInterface:    9000,
		CloudWatchLogGroup:  8900,
		Trail:               8800,
	}

	tagFieldNames = []string{
		"Tags",
		"TagSet",
	}

	// creationTimeFieldNames are a list field names that are used to find the creation date of a resource.
	creationTimeFieldNames = []string{
		"LaunchTime",
		"CreateTime",
		"CreateDate",
		"CreatedTime",
		"CreationDate",
		"CreationTime",
		"CreationTimestamp",
		"StartTime",
		"InstanceCreateTime",
	}
)

func SupportedResourceType(resType TerraformResourceType) bool {
	_, found := deleteIDs[resType]

	return found
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
	autoscalingiface.AutoScalingAPI
	cloudformationiface.CloudFormationAPI
	cloudtrailiface.CloudTrailAPI
	cloudwatchlogsiface.CloudWatchLogsAPI
	ec2iface.EC2API
	ecsiface.ECSAPI
	efsiface.EFSAPI
	elbiface.ELBAPI
	iamiface.IAMAPI
	kmsiface.KMSAPI
	lambdaiface.LambdaAPI
	rdsiface.RDSAPI
	route53iface.Route53API
	s3iface.S3API
	stsiface.STSAPI
}

type LoadBalancerDescription struct {
	elb.LoadBalancerDescription
	Tags []*elb.Tag
}

// NewAWS creates an AWS instance
func NewAWS(s *session.Session) *AWS {
	return &AWS{
		AutoScalingAPI:    autoscaling.New(s),
		CloudFormationAPI: cloudformation.New(s),
		CloudWatchLogsAPI: cloudwatchlogs.New(s),
		EC2API:            ec2.New(s),
		ECSAPI:            ecs.New(s),
		EFSAPI:            efs.New(s),
		ELBAPI:            elb.New(s),
		IAMAPI:            iam.New(s),
		KMSAPI:            kms.New(s),
		LambdaAPI:         lambda.New(s),
		Route53API:        route53.New(s),
		RDSAPI:            rds.New(s),
		S3API:             s3.New(s),
		STSAPI:            sts.New(s),
	}
}

// Resources is a list of AWS resources.
type Resources []*Resource

// Resource contains information about a single AWS resource that can be deleted by Terraform.
type Resource struct {
	Type TerraformResourceType
	// ID by which the resource can be deleted (in some cases the ID is the resource's name, but not always;
	// that's why we need the deleteIDs map)
	ID      string
	Tags    map[string]string
	Created *time.Time
	Attrs   map[string]string
}

// RawResources lists all resources of a particular type
func (a *AWS) RawResources(resType TerraformResourceType) (interface{}, error) {
	switch resType {
	case Ami:
		return a.amis()
	case AutoscalingGroup:
		return a.autoscalingGroups()
	case CloudformationStack:
		return a.cloudFormationStacks()
	case CloudWatchLogGroup:
		return a.cloudWatchLogGroups()
	case EbsSnapshot:
		return a.ebsSnapshots()
	case EbsVolume:
		return a.ebsVolumes()
	case EcsCluster:
		return a.ecsClusters()
	case EfsFileSystem:
		return a.efsFileSystems()
	case Eip:
		return a.eips()
	case Elb:
		return a.elbs()
	case IamGroup:
		return a.iamGroups()
	case IamInstanceProfile:
		return a.iamInstanceProfiles()
	case IamPolicy:
		return a.iamPolicies()
	case IamRole:
		return a.iamRoles()
	case IamUser:
		return a.iamUsers()
	case Instance:
		return a.instances()
	case InternetGateway:
		return a.internetGateways()
	case KeyPair:
		return a.keyPairs()
	case KmsAlias:
		return a.KmsAliases()
	case KmsKey:
		return a.KmsKeys()
	case LambdaFunction:
		return a.lambdaFunctions()
	case LaunchConfiguration:
		return a.launchConfigurations()
	case NatGateway:
		return a.natGateways()
	case NetworkACL:
		return a.networkAcls()
	case NetworkInterface:
		return a.networkInterfaces()
	case DBInstance:
		return a.rdsInstances()
	case Route53Zone:
		return a.route53Zones()
	case RouteTable:
		return a.routeTables()
	case S3Bucket:
		return a.s3Buckets()
	case SecurityGroup:
		return a.SecurityGroup()
	case Subnet:
		return a.subnets()
	case Trail:
		return a.trails()
	case Vpc:
		return a.vpcs()
	case VpcEndpoint:
		return a.vpcEndpoints()
	default:
		return nil, errors.Errorf("unknown or unsupported resource type: %s", resType)
	}
}

func (a *AWS) instances() (interface{}, error) {
	output, err := a.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("instance-state-name"),
				Values: []*string{
					aws.String("pending"), aws.String("running"),
					aws.String("stopping"), aws.String("stopped"),
				},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	var instances []*ec2.Instance
	for _, r := range output.Reservations {
		instances = append(instances, r.Instances...)
	}

	return instances, nil
}

func (a *AWS) keyPairs() (interface{}, error) {
	output, err := a.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{})
	if err != nil {
		return nil, err
	}
	return output.KeyPairs, nil
}

func (a *AWS) elbs() (interface{}, error) {
	output, err := a.ELBAPI.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})
	if err != nil {
		return nil, err
	}
	loadBalancerNames := make([]string, len(output.LoadBalancerDescriptions))
	loadBalancerNameMap := make(map[string]*elb.LoadBalancerDescription)
	for i, desc := range output.LoadBalancerDescriptions {
		loadBalancerNames[i] = *desc.LoadBalancerName
		loadBalancerNameMap[*(desc.LoadBalancerName)] = desc
	}

	tagDescriptions, err := a.findElbTags(loadBalancerNames)
	if err != nil {
		return nil, err
	}
	loadBalancers := make([]*LoadBalancerDescription, len(output.LoadBalancerDescriptions))
	for i, tagDesc := range tagDescriptions {
		loadBalancer := loadBalancerNameMap[*(tagDesc.LoadBalancerName)]
		loadBalancers[i] = &LoadBalancerDescription{*loadBalancer, tagDesc.Tags}
	}
	return loadBalancers, nil
}

func (a *AWS) findElbTags(elbNames []string) ([]*elb.TagDescription, error) {
	var tagDescriptions []*elb.TagDescription
	batchSize := 20
	for i := 0; i < len(elbNames); i += batchSize {
		end := i + batchSize
		if end > len(elbNames) {
			end = len(elbNames)
		}
		awsNames := make([]*string, end-i)
		for i, n := range elbNames[i:end] {
			awsNames[i] = aws.String(n)
		}
		resp, err := a.ELBAPI.DescribeTags(&elb.DescribeTagsInput{LoadBalancerNames: awsNames})
		if err != nil {
			return nil, fmt.Errorf("DescribeTags SDK error: %v", err)
		}
		tagDescriptions = append(tagDescriptions, resp.TagDescriptions...)
	}
	return tagDescriptions, nil
}

func (a *AWS) vpcEndpoints() (interface{}, error) {
	output, err := a.DescribeVpcEndpoints(&ec2.DescribeVpcEndpointsInput{})
	if err != nil {
		return nil, err
	}
	return output.VpcEndpoints, nil
}

// TODO support findTags
func (a *AWS) natGateways() (interface{}, error) {
	output, err := a.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{
		Filter: []*ec2.Filter{
			{
				Name: aws.String("state"),
				Values: []*string{
					aws.String("available"),
				},
			},
		},
	})

	if err != nil {
		return nil, err
	}
	return output.NatGateways, nil
}

func (a *AWS) cloudFormationStacks() (interface{}, error) {
	output, err := a.DescribeStacks(&cloudformation.DescribeStacksInput{})
	if err != nil {
		return nil, err
	}
	return output.Stacks, nil
}

func (a *AWS) cloudWatchLogGroups() (interface{}, error) {
	output, err := a.CloudWatchLogsAPI.DescribeLogGroups(&cloudwatchlogs.DescribeLogGroupsInput{})
	if err != nil {
		return nil, err
	}
	return output.LogGroups, nil
}

func (a *AWS) route53Zones() (interface{}, error) {
	output, err := a.ListHostedZones(&route53.ListHostedZonesInput{})
	if err != nil {
		return nil, err
	}
	return output.HostedZones, nil
}

func (a *AWS) efsFileSystems() (interface{}, error) {
	output, err := a.DescribeFileSystems(&efs.DescribeFileSystemsInput{})
	if err != nil {
		return nil, err
	}
	return output.FileSystems, nil
}

// Elastic network interface (ENI) resource
// sort by owner of the network interface?
// support findTags
// attached to subnet
func (a *AWS) networkInterfaces() (interface{}, error) {
	output, err := a.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{})
	if err != nil {
		return nil, err
	}
	return output.NetworkInterfaces, nil
}

func (a *AWS) eips() (interface{}, error) {
	output, err := a.DescribeAddresses(&ec2.DescribeAddressesInput{})
	if err != nil {
		return nil, err
	}
	return output.Addresses, nil
}

func (a *AWS) internetGateways() (interface{}, error) {
	output, err := a.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{})
	if err != nil {
		return nil, err
	}
	return output.InternetGateways, nil
}

func (a *AWS) subnets() (interface{}, error) {
	output, err := a.DescribeSubnets(&ec2.DescribeSubnetsInput{})
	if err != nil {
		return nil, err
	}
	return output.Subnets, nil
}

func (a *AWS) routeTables() (interface{}, error) {
	output, err := a.DescribeRouteTables(&ec2.DescribeRouteTablesInput{})
	if err != nil {
		return nil, err
	}
	return output.RouteTables, nil
}

func (a *AWS) rdsInstances() (interface{}, error) {
	output, err := a.DescribeDBInstances(&rds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, err
	}
	return output.DBInstances, nil
}

func (a *AWS) SecurityGroup() (interface{}, error) {
	output, err := a.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})
	if err != nil {
		return nil, err
	}
	return output.SecurityGroups, nil
}

func (a *AWS) networkAcls() (interface{}, error) {
	output, err := a.DescribeNetworkAcls(&ec2.DescribeNetworkAclsInput{})
	if err != nil {
		return nil, err
	}
	return output.NetworkAcls, nil
}

func (a *AWS) vpcs() (interface{}, error) {
	output, err := a.DescribeVpcs(&ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, err
	}
	return output.Vpcs, nil
}

func (a *AWS) iamPolicies() (interface{}, error) {
	output, err := a.ListPolicies(&iam.ListPoliciesInput{})
	if err != nil {
		return nil, err
	}
	return output.Policies, nil
}

func (a *AWS) iamGroups() (interface{}, error) {
	output, err := a.ListGroups(&iam.ListGroupsInput{})
	if err != nil {
		return nil, err
	}
	return output.Groups, nil
}

func (a *AWS) iamUsers() (interface{}, error) {
	output, err := a.ListUsers(&iam.ListUsersInput{})
	if err != nil {
		return nil, err
	}
	return output.Users, nil
}

func (a *AWS) iamRoles() (interface{}, error) {
	output, err := a.ListRoles(&iam.ListRolesInput{})
	if err != nil {
		return nil, err
	}
	return output.Roles, nil
}

func (a *AWS) iamInstanceProfiles() (interface{}, error) {
	output, err := a.ListInstanceProfiles(&iam.ListInstanceProfilesInput{})
	if err != nil {
		return nil, err
	}
	return output.InstanceProfiles, nil
}

func (a *AWS) KmsAliases() (interface{}, error) {
	output, err := a.KMSAPI.ListAliases(&kms.ListAliasesInput{})
	if err != nil {
		return nil, err
	}
	return output.Aliases, nil
}

func (a *AWS) KmsKeys() (interface{}, error) {
	output, err := a.ListKeys(&kms.ListKeysInput{})
	if err != nil {
		return nil, err
	}
	return output.Keys, nil
}

func (a *AWS) s3Buckets() (interface{}, error) {
	output, err := a.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}
	return output.Buckets, nil
}

func (a *AWS) trails() (interface{}, error) {
	output, err := a.ListTrails(&cloudtrail.ListTrailsInput{})
	if err != nil {
		return nil, err
	}
	return output.Trails, nil

func (a *AWS) ebsSnapshots() (interface{}, error) {
	output, err := a.DescribeSnapshots(&ec2.DescribeSnapshotsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("owner-id"),
				Values: []*string{
					a.callerIdentity(),
				},
			},
		},
	})

	if err != nil {
		return nil, err
	}
	return output.Snapshots, nil
}

func (a *AWS) ebsVolumes() (interface{}, error) {
	output, err := a.DescribeVolumes(&ec2.DescribeVolumesInput{})
	if err != nil {
		return nil, err
	}
	return output.Volumes, nil
}

func (a *AWS) ecsClusters() (interface{}, error) {
	listOutput, err := a.ListClusters(&ecs.ListClustersInput{})
	if err != nil {
		return nil, err
	}

	descOutput, err := a.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: listOutput.ClusterArns,
		Include:  []*string{aws.String("TAGS")},
	})

	return descOutput.Clusters, nil
}

func (a *AWS) amis() (interface{}, error) {
	output, err := a.DescribeImages(&ec2.DescribeImagesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("owner-id"),
				Values: []*string{
					a.callerIdentity(),
				},
			},
		},
	})

	if err != nil {
		return nil, err
	}
	return output.Images, nil
}

func (a *AWS) autoscalingGroups() (interface{}, error) {
	output, err := a.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	if err != nil {
		return nil, err
	}
	return output.AutoScalingGroups, nil
}

func (a *AWS) lambdaFunctions() (interface{}, error) {
	output, err := a.ListFunctions(&lambda.ListFunctionsInput{})
	if err != nil {
		return nil, err
	}
	return output.Functions, nil
}

func (a *AWS) launchConfigurations() (interface{}, error) {
	output, err := a.DescribeLaunchConfigurations(&autoscaling.DescribeLaunchConfigurationsInput{})
	if err != nil {
		return nil, err
	}
	return output.LaunchConfigurations, nil
}

// callerIdentity returns the account ID of the AWS account for the currently used credentials
func (a *AWS) callerIdentity() *string {
	res, err := a.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		log.Fatal(err)
	}
	return res.Account
}
