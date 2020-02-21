module github.com/cloudetc/awsweeper

require (
	github.com/apex/log v1.1.2
	github.com/aws/aws-sdk-go v1.29.1
	github.com/fatih/color v1.9.0
	github.com/go-errors/errors v1.0.2-0.20180813162953-d98b870cc4e0
	github.com/golang/mock v1.4.0
	github.com/gruntwork-io/terratest v0.23.0
	github.com/hashicorp/hcl v0.0.0-20171017181929-23c074d0eceb // indirect
	github.com/hashicorp/terraform v0.12.20
	github.com/hashicorp/yamux v0.0.0-20181012175058-2f1d1f20f75d // indirect
	github.com/jckuester/terradozer v0.0.0-20200220204151-8622c32b5cbc
	github.com/mitchellh/cli v1.0.0
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/onsi/gomega v1.9.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/afero v1.2.2 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/tj/assert v0.0.0-20171129193455-018094318fb0
	golang.org/x/crypto v0.0.0-20190829043050-9756ffdc2472 // indirect
	gopkg.in/yaml.v2 v2.2.4
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

go 1.13
