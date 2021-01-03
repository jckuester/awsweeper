package resource

var (
	// DependencyOrder is the order in which resource types should be deleted,
	// since dependent resources need to be deleted before their dependencies
	// (e.g. aws_subnet before aws_vpc)
	DependencyOrder = map[string]int{
		"aws_lambda_function":      10100,
		"aws_ecs_cluster":          10000,
		"aws_autoscaling_group":    9990,
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
		"aws_ami":                  9740,
		"aws_ebs_volume":           9730,
		"aws_ebs_snapshot":         9720,
		"aws_kms_alias":            9610,
		"aws_kms_key":              9600,
		"aws_network_interface":    9000,
		"aws_cloudwatch_log_group": 8900,
		"aws_cloudtrail":           8800,
	}
)
