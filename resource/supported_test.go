package resource

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSupported(t *testing.T) {
	apiInfo, err := getSupported("aws_vpc", &AWSClient{})

	require.Equal(t, apiInfo.TerraformType, "aws_vpc")
	require.Equal(t, apiInfo.DeleteId, "VpcId")

	require.NoError(t, err)
}

func TestGetSupported_InvalidType(t *testing.T) {
	_, err := getSupported("some_type", &AWSClient{})

	require.Error(t, err)
}
