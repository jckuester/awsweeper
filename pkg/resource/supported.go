package resource

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"

	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awsls "github.com/jckuester/awsls/aws"
	"github.com/pkg/errors"
)

const (
	Ami              = "aws_ami"
	AutoscalingGroup = "aws_autoscaling_group"
	EbsSnapshot      = "aws_ebs_snapshot"
	EcsCluster       = "aws_ecs_cluster"
	CloudTrail       = "aws_cloudtrail"
)

var (
	deleteIDs = map[string]string{
		Ami:              "ImageId",
		AutoscalingGroup: "AutoScalingGroupName",
		EbsSnapshot:      "SnapshotId",
		// Note: to import a cluster, the name is used as ID
		EcsCluster: "ClusterArn",
		CloudTrail: "Name",
	}

	// DependencyOrder is the order in which resource types should be deleted,
	// since dependent resources need to be deleted before their dependencies
	// (e.g. aws_subnet before aws_vpc)
	DependencyOrder = map[string]int{
		"aws_lambda_function":      10100,
		"aws_ecs_cluster":          10000,
		AutoscalingGroup:           9990,
		"aws_instance":             9980,
		"aws_key_pair":             9970,
		"aws_elb":                  9960,
		"aws_vpc_endpoint":         9950,
		"aws_nat_gateway":          9940,
		"aws_cloudformation_stack": 9930,
		"aws_route53_zone":         9920,
		"aws_efs_file_system":      9910,
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
		"aws_kms_alias":            9610,
		"aws_kms_key":              9600,
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
type AWS awsls.Client

// RawResources lists all resources of a particular type
func (a AWS) RawResources(resType string) (interface{}, error) {
	switch resType {
	case Ami:
		return a.amis()
	case AutoscalingGroup:
		return a.autoscalingGroups()
	case EbsSnapshot:
		return a.ebsSnapshots()
	case EcsCluster:
		return a.ecsClusters()
	case CloudTrail:
		return a.cloudTrails()
	default:
		return nil, errors.Errorf("unknown or unsupported resource type: %s", resType)
	}
}

func (a *AWS) ecsClusters() (interface{}, error) {
	listClustersRequest := a.Ecsconn.ListClustersRequest(&ecs.ListClustersInput{})

	var clusterARNs []string

	pg := ecs.NewListClustersPaginator(listClustersRequest)
	for pg.Next(context.Background()) {
		page := pg.CurrentPage()

		clusterARNs = append(clusterARNs, page.ClusterArns...)
	}

	if err := pg.Err(); err != nil {
		return nil, err
	}

	// TODO is paginated, but not paginator API
	req := a.Ecsconn.DescribeClustersRequest(&ecs.DescribeClustersInput{
		Clusters: clusterARNs,
		Include:  []ecs.ClusterField{"TAGS"},
	})

	resp, err := req.Send(context.Background())
	if err != nil {
		return nil, err
	}

	return resp.Clusters, nil
}

func (a *AWS) cloudTrails() (interface{}, error) {
	req := a.Cloudtrailconn.DescribeTrailsRequest(&cloudtrail.DescribeTrailsInput{})

	resp, err := req.Send(context.Background())
	if err != nil {
		return nil, err
	}

	return resp.TrailList, nil
}

func (a *AWS) ebsSnapshots() (interface{}, error) {
	req := a.Ec2conn.DescribeSnapshotsRequest(&ec2.DescribeSnapshotsInput{
		Filters: []ec2.Filter{
			{
				Name: aws.String("owner-id"),
				Values: []string{
					a.AccountID,
				},
			},
		},
	})

	var snapshots []ec2.Snapshot

	pg := ec2.NewDescribeSnapshotsPaginator(req)
	for pg.Next(context.Background()) {
		page := pg.CurrentPage()

		snapshots = append(snapshots, page.Snapshots...)
	}

	return snapshots, nil
}

func (a *AWS) amis() (interface{}, error) {
	req := a.Ec2conn.DescribeImagesRequest(&ec2.DescribeImagesInput{
		Filters: []ec2.Filter{
			{
				Name: aws.String("owner-id"),
				Values: []string{
					a.AccountID,
				},
			},
		},
	})

	resp, err := req.Send(context.Background())
	if err != nil {
		return nil, err
	}

	return resp.Images, nil
}

func (a *AWS) autoscalingGroups() (interface{}, error) {
	req := a.Autoscalingconn.DescribeAutoScalingGroupsRequest(&autoscaling.DescribeAutoScalingGroupsInput{})

	var autoScalingGroups []autoscaling.AutoScalingGroup

	pg := autoscaling.NewDescribeAutoScalingGroupsPaginator(req)
	for pg.Next(context.Background()) {
		page := pg.CurrentPage()

		autoScalingGroups = append(autoScalingGroups, page.AutoScalingGroups...)
	}

	if err := pg.Err(); err != nil {
		return nil, err
	}

	return autoScalingGroups, nil
}
