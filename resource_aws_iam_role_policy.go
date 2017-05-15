package main

import (
	"github.com/aws/aws-sdk-go/service/iam"
	"fmt"
	"github.com/jckuester/awsweeper/schema"
	"github.com/aws/aws-sdk-go/aws"
	"strings"
)

func resourceAwsIamRolePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	role, name, err := resourceAwsIamRolePolicyParseId(d.Id())
	if err != nil {
		return err
	}

	request := &iam.DeleteRolePolicyInput{
		PolicyName: aws.String(name),
		RoleName:   aws.String(role),
	}

	if _, err := iamconn.DeleteRolePolicy(request); err != nil {
		return fmt.Errorf("Error deleting IAM role policy %s: %s", d.Id(), err)
	}
	return nil
}

func resourceAwsIamRolePolicyParseId(id string) (roleName, policyName string, err error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("role_policy id must be of the form <role name>:<policy name>")
		return
	}

	roleName = parts[0]
	policyName = parts[1]
	return
}
