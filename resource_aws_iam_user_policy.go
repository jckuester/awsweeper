package main

import (
	"github.com/aws/aws-sdk-go/service/iam"
	"fmt"
	"github.com/jckuester/awsweeper/schema"
	"github.com/aws/aws-sdk-go/aws"
	"strings"
)

func resourceAwsIamUserPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	user, name := resourceAwsIamUserPolicyParseId(d.Id())

	request := &iam.DeleteUserPolicyInput{
		PolicyName: aws.String(name),
		UserName:   aws.String(user),
	}

	if _, err := iamconn.DeleteUserPolicy(request); err != nil {
		return fmt.Errorf("Error deleting IAM user policy %s: %s", d.Id(), err)
	}
	return nil
}

func resourceAwsIamUserPolicyParseId(id string) (userName, policyName string) {
	parts := strings.SplitN(id, ":", 2)
	userName = parts[0]
	policyName = parts[1]
	return
}
