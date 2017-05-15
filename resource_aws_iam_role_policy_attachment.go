package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/jckuester/awsweeper/schema"
	"github.com/aws/aws-sdk-go/aws"
)

func resourceAwsIamRolePolicyAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn
	role := d.Get("role").(string)
	arn := d.Get("policy_arn").(string)

	err := detachPolicyFromRole(conn, role, arn)
	if err != nil {
		return fmt.Errorf("[WARN] Error removing policy %s from IAM Role %s: %v", arn, role, err)
	}
	return nil
}

func detachPolicyFromRole(conn *iam.IAM, role string, arn string) error {
	_, err := conn.DetachRolePolicy(&iam.DetachRolePolicyInput{
		RoleName:  aws.String(role),
		PolicyArn: aws.String(arn),
	})
	if err != nil {
		return err
	}
	return nil
}
