module github.com/jckuester/awsweeper

go 1.17

require (
	github.com/apex/log v1.9.0
	github.com/aws/aws-sdk-go-v2 v1.6.0
	github.com/aws/aws-sdk-go-v2/config v1.1.1
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.1.1
	github.com/aws/aws-sdk-go-v2/service/cloudtrail v1.1.1
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.1.1
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.1.1
	github.com/aws/aws-sdk-go-v2/service/ecs v1.1.1
	github.com/aws/aws-sdk-go-v2/service/efs v1.1.1
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing v1.1.1
	github.com/aws/aws-sdk-go-v2/service/iam v1.1.1
	github.com/aws/aws-sdk-go-v2/service/lambda v1.1.1
	github.com/aws/aws-sdk-go-v2/service/rds v1.1.1
	github.com/aws/smithy-go v1.9.1
	github.com/fatih/color v1.10.0
	github.com/gruntwork-io/terratest v0.24.2
	github.com/jckuester/awsls v0.11.1-0.20211024194801-688a8938b1b3
	github.com/jckuester/awstools-lib v0.0.0-20210524191941-23f0e367139d
	github.com/jckuester/terradozer v0.1.4-0.20210524190016-3e6d42479316
	github.com/onsi/gomega v1.9.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/zclconf/go-cty v1.7.1
	gopkg.in/yaml.v2 v2.2.8
)

require (
	cloud.google.com/go v0.51.0 // indirect
	cloud.google.com/go/storage v1.0.0 // indirect
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/agext/levenshtein v1.2.2 // indirect
	github.com/apparentlymart/go-cidr v1.0.1 // indirect
	github.com/apparentlymart/go-textseg v1.0.0 // indirect
	github.com/apparentlymart/go-textseg/v12 v12.0.0 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-sdk-go v1.38.43 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.0.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/accessanalyzer v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/acm v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/acmpca v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/apigateway v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/apigatewayv2 v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/applicationautoscaling v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/appmesh v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/apprunner v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/appsync v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/athena v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/autoscalingplans v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/backup v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/batch v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/budgets v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/cloud9 v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/cloudformation v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/cloudhsmv2 v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/cloudwatchevents v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/codeartifact v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/codebuild v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/codecommit v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/codedeploy v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/codepipeline v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/codestarconnections v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/codestarnotifications v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/cognitoidentity v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/configservice v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/costandusagereportservice v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/databasemigrationservice v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/datapipeline v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/datasync v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/dax v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/devicefarm v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/directconnect v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/directoryservice v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/dlm v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/docdb v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ecr v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ecrpublic v1.1.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/eks v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/elasticache v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/elasticsearchservice v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/elastictranscoder v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/emr v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/firehose v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/fms v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/fsx v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/gamelift v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/glacier v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/globalaccelerator v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/glue v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/guardduty v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/imagebuilder v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/inspector v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.0.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.0.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.1.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/iot v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/kafka v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/kinesisanalytics v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2 v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/kinesisvideo v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/kms v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/lakeformation v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/licensemanager v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/lightsail v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/macie v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/macie2 v1.6.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/mediaconvert v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/mediapackage v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/mediastore v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/mq v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/mwaa v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/neptune v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/networkfirewall v1.1.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/opsworks v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/organizations v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/pinpoint v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/qldb v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/quicksight v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ram v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/redshift v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/resourcegroups v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/route53 v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/route53resolver v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.2.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3control v1.2.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3outposts v1.1.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sagemaker v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/securityhub v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/servicecatalog v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/servicediscovery v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/servicequotas v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ses v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sfn v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/shield v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/signer v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sns v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sqs v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssm v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssoadmin v1.1.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/storagegateway v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/swf v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/synthetics v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/timestreamwrite v1.2.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/transfer v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/waf v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/wafregional v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/wafv2 v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/worklink v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/workspaces v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/xray v1.1.1 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/bgentry/speakeasy v0.1.0 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/bmatcuk/doublestar v1.1.5 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/groupcache v0.0.0-20191227052852-215e87163ea7 // indirect
	github.com/golang/protobuf v1.3.4 // indirect
	github.com/google/go-cmp v0.5.4 // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/googleapis/gax-go/v2 v2.0.5 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/hashicorp/go-getter v1.4.2-0.20200106182914-9813cbd4eb02 // indirect
	github.com/hashicorp/go-hclog v0.12.0 // indirect
	github.com/hashicorp/go-multierror v1.0.0 // indirect
	github.com/hashicorp/go-plugin v1.3.0 // indirect
	github.com/hashicorp/go-retryablehttp v0.5.2 // indirect
	github.com/hashicorp/go-safetemp v1.0.0 // indirect
	github.com/hashicorp/go-uuid v1.0.1 // indirect
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/hcl/v2 v2.3.0 // indirect
	github.com/hashicorp/hil v0.0.0-20190212112733-ab17b08d6590 // indirect
	github.com/hashicorp/terraform v0.12.31 // indirect
	github.com/hashicorp/terraform-config-inspect v0.0.0-20191212124732-c6ae6269b9d7 // indirect
	github.com/hashicorp/terraform-svchost v0.0.0-20191011084731-65d371908596 // indirect
	github.com/hashicorp/yamux v0.0.0-20180604194846-3520598351bb // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jstemmer/go-junit-report v0.9.1 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mitchellh/cli v1.0.0 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-testing-interface v1.0.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/hashstructure v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/mitchellh/reflectwalk v1.0.0 // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/posener/complete v1.2.1 // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/ulikunitz/xz v0.5.5 // indirect
	github.com/vmihailenco/msgpack/v4 v4.3.12 // indirect
	github.com/vmihailenco/tagparser v0.1.1 // indirect
	github.com/zclconf/go-cty-yaml v1.0.1 // indirect
	go.opencensus.io v0.22.2 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/exp v0.0.0-20191227195350-da58074b4299 // indirect
	golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f // indirect
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20200930185726-fdedc70b468f // indirect
	golang.org/x/text v0.3.3 // indirect
	golang.org/x/tools v0.0.0-20200113040837-eac381796e91 // indirect
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543 // indirect
	google.golang.org/api v0.15.0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/genproto v0.0.0-20200108215221-bd8f9a0ef82f // indirect
	google.golang.org/grpc v1.27.1 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200605160147-a5ece683394c // indirect
	honnef.co/go/tools v0.0.1-2019.2.3 // indirect
)
