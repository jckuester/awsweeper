package resource

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/require"
)

var (
	securityGroupType = SecurityGroup
	iamRoleType       = IamRole
	instanceType      = Instance
	vpc               = Vpc

	yml = YamlCfg{
		iamRoleType: {
			Ids: []*string{aws.String("^foo.*")},
		},
		securityGroupType: {},
		instanceType: {
			Tags: map[string]string{
				"foo": "bar",
				"bla": "blub",
			},
		},
		vpc: {
			Ids: []*string{aws.String("^foo.*")},
			Tags: map[string]string{
				"foo": "bar",
			},
		},
	}

	f = &YamlFilter{
		cfg: yml,
	}
)

//func TestValidate(t *testing.T) {
//	apiDescs := Supported(mockAWSClient())
//
//	require.NoError(t, f.Validate(apiDescs))
//}
//
//func TestValidate_EmptyCfg(t *testing.T) {
//	apiDescs := Supported(mockAWSClient())
//
//	require.NoError(t, f.Validate(apiDescs))
//}
//
//func TestValidate_NotSupportedResourceTypeInCfg(t *testing.T) {
//	apiDescs := Supported(mockAWSClient())
//
//	f := &YamlFilter{
//		cfg: YamlCfg{
//			securityGroupType:    {},
//			"not_supported_type": {},
//		},
//	}
//
//	require.Error(t, f.Validate(apiDescs))
//}

func TestResourceTypes(t *testing.T) {
	resTypes := f.Types()

	require.Len(t, resTypes, len(yml))
	require.Contains(t, resTypes, securityGroupType)
	require.Contains(t, resTypes, iamRoleType)
	require.Contains(t, resTypes, instanceType)
}

func TestResourceTypes_emptyCfg(t *testing.T) {
	rf := &YamlFilter{
		cfg: YamlCfg{},
	}

	resTypes := rf.Types()

	require.Len(t, resTypes, 0)
	require.Empty(t, resTypes)
}

func TestResourceMatchIds_IdMatchesFilterCriteria(t *testing.T) {
	matchesID, err := f.matchID(iamRoleType, "foo-lala")

	require.True(t, matchesID)
	require.NoError(t, err)
}

func TestResourceMatchIds_IdDoesNotMatchFilterCriteria(t *testing.T) {
	matchesID, err := f.matchID(iamRoleType, "lala-foo")

	require.False(t, matchesID)
	require.NoError(t, err)
}

func TestResourceMatchIds_NoFilterCriteriaSetForIds(t *testing.T) {
	_, err := f.matchID(securityGroupType, "matches-any-id")

	require.Error(t, err)
}

func TestResourceMatchTags_TagMatchesFilterCriteria(t *testing.T) {
	matchesTags, err := f.matchTags(instanceType, map[string]string{"foo": "bar"})

	require.True(t, matchesTags)
	require.NoError(t, err)

	matchesTags, err = f.matchTags(instanceType, map[string]string{"bla": "blub"})

	require.True(t, matchesTags)
	require.NoError(t, err)
}

func TestResourceMatchTags_TagDoesNotMatchFilterCriteria(t *testing.T) {
	matchesTags, err := f.matchTags(instanceType, map[string]string{"foo": "baz"})

	require.False(t, matchesTags)
	require.NoError(t, err)

	matchesTags, err = f.matchTags(instanceType, map[string]string{"blub": "bla"})

	require.False(t, matchesTags)
	require.NoError(t, err)
}

func TestResourceMatchTags_NoFilterCriteriaSetForTags(t *testing.T) {
	_, err := f.matchTags(securityGroupType, map[string]string{"any": "tag"})

	require.Error(t, err)
}

func TestMatch_OnlyTagFilterCriteria(t *testing.T) {
	require.True(t, f.Matches(instanceType, "foo-lala", map[string]string{"foo": "bar"}))
	require.False(t, f.Matches(instanceType, "some-id", map[string]string{"any": "tag"}))
	require.False(t, f.Matches(instanceType, "some-id"))
}

func TestMatch_OnlyIdFilterCriteria(t *testing.T) {
	require.True(t, f.Matches(iamRoleType, "foo-lala", map[string]string{"any": "tag"}))
	require.False(t, f.Matches(iamRoleType, "some-id", map[string]string{"foo": "bar"}))
	require.False(t, f.Matches(iamRoleType, "some-id"))
}

func TestMatch_IdAndTagFilterCriteria(t *testing.T) {
	require.True(t, f.Matches(vpc, "foo-lala", map[string]string{"any": "tag"}))
	require.True(t, f.Matches(vpc, "some-id", map[string]string{"foo": "bar"}))
	require.False(t, f.Matches(vpc, "some-id", map[string]string{"any": "tag"}))
}

func TestMatch_NoFilterCriteriaGiven(t *testing.T) {
	require.True(t, f.Matches(securityGroupType, "any-id", map[string]string{"any": "tag"}))
}
