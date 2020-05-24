package resource

import (
	"fmt"
	"strings"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/kms"
	awsls "github.com/jckuester/awsls/aws"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// here is where the filtering of resources happens, i.e.
// the filter entry in the config for a certain resource type
// is applied to all resources of that type.
func (f Filter) Apply(resType string, res []awsls.Resource, raw interface{}, aws *AWS) []awsls.Resource {
	for i, r := range res {
		tags, err := GetTags(&r)
		if err != nil {
			log.WithFields(log.Fields{
				"type": r.Type,
				"id":   r.ID,
			}).WithError(err).Debug("failed to get tags")

			continue
		}

		res[i].Tags = tags
	}

	switch resType {
	case EfsFileSystem:
		return f.efsFileSystemFilter(res, raw, aws)
	case KmsKey:
		return f.kmsKeysFilter(res, aws)
	case KmsAlias:
		return f.kmsKeyAliasFilter(res)
	default:
		return f.defaultFilter(res)
	}
}

func GetTags(r *awsls.Resource) (map[string]string, error) {
	if r.Resource == nil {
		return nil, fmt.Errorf("resource is nil")
	}

	state := r.State()

	if state == nil {
		return nil, fmt.Errorf("state is nil")
	}

	if !state.CanIterateElements() {
		return nil, fmt.Errorf("cannot iterate: %s", *state)
	}

	attrValue, ok := state.AsValueMap()["tags"]
	if !ok {
		return nil, fmt.Errorf("attribute not found: tags")
	}

	switch attrValue.Type() {
	case cty.Map(cty.String):
		var v map[string]string
		err := gocty.FromCtyValue(attrValue, &v)
		if err != nil {
			return nil, err
		}

		return v, nil
	default:
		return nil, fmt.Errorf("currently unhandled type: %s", attrValue.Type().FriendlyName())
	}
}

// For most resource types, this default filter method can be used.
// However, for some resource types additional information need to be queried from the AWS API. Filtering for those
// is handled in special functions below.
func (f Filter) defaultFilter(res []awsls.Resource) []awsls.Resource {
	var result []awsls.Resource

	for _, r := range res {
		if f.Match(r) {
			result = append(result, r)
		}
	}
	return result
}

func (f Filter) efsFileSystemFilter(res []awsls.Resource, raw interface{}, c *AWS) []awsls.Resource {
	var result []awsls.Resource

	for _, r := range res {
		if f.Match(awsls.Resource{Type: r.Type, ID: *raw.([]*efs.FileSystemDescription)[0].Name}) {
			res, err := c.DescribeMountTargets(&efs.DescribeMountTargetsInput{
				FileSystemId: &r.ID,
			})

			if err == nil {
				for _, r := range res.MountTargets {
					result = append(result, awsls.Resource{
						Type: "aws_efs_mount_target",
						ID:   *r.MountTargetId,
					})
				}
			}
			result = append(result, r)
		}
	}

	return result
}

func (f Filter) kmsKeysFilter(res []awsls.Resource, c *AWS) []awsls.Resource {
	var result []awsls.Resource

	for _, r := range res {
		if f.Match(r) {
			req, res := c.DescribeKeyRequest(&kms.DescribeKeyInput{
				KeyId: aws.String(r.ID),
			})
			err := req.Send()
			if err == nil {
				if *res.KeyMetadata.KeyState != "PendingDeletion" {
					result = append(result, awsls.Resource{
						Type: "aws_kms_key",
						ID:   r.ID,
					})
				}
			}
		}
	}

	// associated aliases will also be deleted after waiting period (between 7 to 30 days)
	return result
}

func (f Filter) kmsKeyAliasFilter(res []awsls.Resource) []awsls.Resource {
	var result []awsls.Resource

	for _, r := range res {
		if f.Match(r) && !strings.HasPrefix(r.ID, "alias/aws/") {
			result = append(result, r)
		}
	}

	return result
}
