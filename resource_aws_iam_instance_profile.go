package main

import (
	"github.com/aws/aws-sdk-go/service/iam"
	"fmt"
	"github.com/jckuester/awsweeper/schema"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
)

func resourceAwsIamInstanceProfileDelete(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	if err := instanceProfileRemoveAllRoles(d, iamconn); err != nil {
		return err
	}

	request := &iam.DeleteInstanceProfileInput{
		InstanceProfileName: aws.String(d.Id()),
	}
	_, err := iamconn.DeleteInstanceProfile(request)
	if err != nil {
		return fmt.Errorf("Error deleting IAM instance profile %s: %s", d.Id(), err)
	}
	d.SetId("")
	return nil
}

func instanceProfileRemoveAllRoles(d *schema.ResourceData, iamconn *iam.IAM) error {
	for _, role := range d.Get("roles").(*schema.Set).List() {
		err := instanceProfileRemoveRole(iamconn, d.Id(), role.(string))
		if err != nil {
			return fmt.Errorf("Error removing role %s from IAM instance profile %s: %s", role, d.Id(), err)
		}
	}
	return nil
}

func instanceProfileRemoveRole(iamconn *iam.IAM, profileName, roleName string) error {
	request := &iam.RemoveRoleFromInstanceProfileInput{
		InstanceProfileName: aws.String(profileName),
		RoleName:            aws.String(roleName),
	}

	_, err := iamconn.RemoveRoleFromInstanceProfile(request)
	if iamerr, ok := err.(awserr.Error); ok && iamerr.Code() == "NoSuchEntity" {
		return nil
	}
	return err
}
