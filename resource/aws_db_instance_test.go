package resource_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"

	"github.com/cloudetc/awsweeper/resource/mocks"

	"github.com/cloudetc/awsweeper/resource"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAWS_RawResources_DBInstances(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mock := mocks.NewMockRDSAPI(mockCtrl)

	awsClient := &resource.AWS{
		RDSAPI: mock,
	}

	// given
	expectedID := "foo"

	mock.EXPECT().DescribeDBInstances(gomock.Any()).Return(
		&rds.DescribeDBInstancesOutput{
			DBInstances: []*rds.DBInstance{
				{
					DBInstanceIdentifier: &expectedID,
				},
			},
		}, nil)

	// when
	actualResources, err := awsClient.RawResources(resource.DBInstance)
	require.NoError(t, err)
	res := actualResources.([]*rds.DBInstance)

	// then
	assert.Len(t, res, 1)
	assert.Equal(t, *res[0].DBInstanceIdentifier, expectedID)
}
