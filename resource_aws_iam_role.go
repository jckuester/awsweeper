package main

import (
	"github.com/aws/aws-sdk-go/service/iam"
	"fmt"
	"github.com/jckuester/awsweeper/schema"
	"github.com/aws/aws-sdk-go/aws"
)

func resourceAwsIamRoleDelete(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	// Roles cannot be destroyed when attached to an existing Instance Profile
	resp, err := iamconn.ListInstanceProfilesForRole(&iam.ListInstanceProfilesForRoleInput{
		RoleName: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("Error listing Profiles for IAM Role (%s) when trying to delete: %s", d.Id(), err)
	}

	// Loop and remove this Role from any Profiles
	if len(resp.InstanceProfiles) > 0 {
		for _, i := range resp.InstanceProfiles {
			_, err := iamconn.RemoveRoleFromInstanceProfile(&iam.RemoveRoleFromInstanceProfileInput{
				InstanceProfileName: i.InstanceProfileName,
				RoleName:            aws.String(d.Id()),
			})
			if err != nil {
				return fmt.Errorf("Error deleting IAM Role %s: %s", d.Id(), err)
			}
		}
	}

	request := &iam.DeleteRoleInput{
		RoleName: aws.String(d.Id()),
	}

	if _, err := iamconn.DeleteRole(request); err != nil {
		return fmt.Errorf("Error deleting IAM Role %s: %s", d.Id(), err)
	}
	return nil
}
