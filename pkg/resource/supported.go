package resource

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/aws/aws-sdk-go/service/cloudtrail/cloudtrailiface"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
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

const (
	Ami              = "aws_ami"
	AutoscalingGroup = "aws_autoscaling_group"
	EbsSnapshot      = "aws_ebs_snapshot"
	EcsCluster       = "aws_ecs_cluster"
	EfsFileSystem    = "aws_efs_file_system"
	Instance         = "aws_instance"
	KmsAlias         = "aws_kms_alias"
	KmsKey           = "aws_kms_key"
	NatGateway       = "aws_nat_gateway"
	CloudTrail       = "aws_cloudtrail"
)

var (
	deleteIDs = map[string]string{
		Ami:              "ImageId",
		AutoscalingGroup: "AutoScalingGroupName",
		// Note: to import a cluster, the name is used as ID
		EcsCluster:    "ClusterArn",
		EfsFileSystem: "FileSystemId",
		Instance:      "InstanceId",
		KmsAlias:      "AliasName",
		KmsKey:        "KeyId",
		NatGateway:    "NatGatewayId",
		CloudTrail:    "Name",
	}

	// DependencyOrder is the order in which resource types should be deleted,
	// since dependent resources need to be deleted before their dependencies
	// (e.g. aws_subnet before aws_vpc)
	DependencyOrder = map[string]int{
		"aws_lambda_function":      10100,
		"aws_ecs_cluster":          10000,
		AutoscalingGroup:           9990,
		Instance:                   9980,
		"aws_key_pair":             9970,
		"aws_elb":                  9960,
		"aws_vpc_endpoint":         9950,
		NatGateway:                 9940,
		"aws_cloudformation_stack": 9930,
		"aws_route53_zone":         9920,
		EfsFileSystem:              9910,
		"aws_launch_configuration": 9900,
		"aws_eip":                  9890,
		"aws_internet_gateway":     9880,
		"aws_subnet":               9870,
		"aws_route_table":          9860,
		"aws_security_group":       9850,
		"aws_network_acl":          9840,
		"aws_vpc":                  9830,
		"aws_db_instance":          9825,
		"aws_iam_policy":           9820,
		"aws_iam_group":            9810,
		"aws_iam_user":             9800,
		"aws_iam_role":             9790,
		"aws_iam_instance_profile": 9780,
		"aws_s3_bucket":            9750,
		Ami:                        9740,
		"aws_ebs_volume":           9730,
		EbsSnapshot:                9720,
		KmsAlias:                   9610,
		KmsKey:                     9600,
		"aws_network_interface":    9000,
		"aws_cloudwatch_log_group": 8900,
		CloudTrail:                 8800,
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

func SupportedResourceType(resType string) bool {
	_, found := deleteIDs[resType]

	return found
}

func getDeleteID(resType string) (string, error) {
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

// NewAWS creates an AWS instance
func NewAWS(s *session.Session) *AWS {
	return &AWS{
		AutoScalingAPI:    autoscaling.New(s),
		CloudFormationAPI: cloudformation.New(s),
		CloudTrailAPI:     cloudtrail.New(s),
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

// RawResources lists all resources of a particular type
func (a *AWS) RawResources(resType string) (interface{}, error) {
	switch resType {
	case Ami:
		return a.amis()
	case AutoscalingGroup:
		return a.autoscalingGroups()
	case EbsSnapshot:
		return a.ebsSnapshots()
	case EcsCluster:
		return a.ecsClusters()
	case EfsFileSystem:
		return a.efsFileSystems()
	case Instance:
		return a.instances()
	case KmsAlias:
		return a.KmsAliases()
	case KmsKey:
		return a.KmsKeys()
	case NatGateway:
		return a.natGateways()
	case CloudTrail:
		return a.cloudTrails()
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

func (a *AWS) efsFileSystems() (interface{}, error) {
	output, err := a.DescribeFileSystems(&efs.DescribeFileSystemsInput{})
	if err != nil {
		return nil, err
	}
	return output.FileSystems, nil
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

func (a *AWS) cloudTrails() (interface{}, error) {
	output, err := a.DescribeTrails(&cloudtrail.DescribeTrailsInput{})
	if err != nil {
		return nil, err
	}
	return output.TrailList, nil
}

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

// callerIdentity returns the account ID of the AWS account for the currently used credentials
func (a *AWS) callerIdentity() *string {
	res, err := a.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		log.Fatal(err)
	}
	return res.Account
}
