module github.com/cloudetc/awsweeper

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

go 1.13

require (
	github.com/apex/log v1.1.2
	github.com/aws/aws-sdk-go v1.29.3
	github.com/go-errors/errors v1.0.2-0.20180813162953-d98b870cc4e0
	github.com/golang/mock v1.4.0
	github.com/hashicorp/terraform v0.12.20
	github.com/jckuester/terradozer v0.0.0-20200216204332-c83267564339
	github.com/mitchellh/cli v1.0.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/afero v1.2.2
	github.com/stretchr/testify v1.4.0
	github.com/terraform-providers/terraform-provider-aws v1.60.0
	golang.org/x/tools v0.0.0-20200216192241-b320d3a0f5a2 // indirect
	gopkg.in/yaml.v2 v2.2.8
)
