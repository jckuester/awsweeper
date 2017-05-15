package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/jckuester/awsweeper/schema"
	"github.com/aws/aws-sdk-go/aws"
)

func resourceAwsIamUserPolicyAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn
	user := d.Get("user").(string)
	arn := d.Get("policy_arn").(string)

	err := detachPolicyFromUser(conn, user, arn)
	if err != nil {
		return fmt.Errorf("[WARN] Error removing policy %s from IAM User %s: %v", arn, user, err)
	}
	return nil
}

func detachPolicyFromUser(conn *iam.IAM, user string, arn string) error {
	_, err := conn.DetachUserPolicy(&iam.DetachUserPolicyInput{
		UserName:  aws.String(user),
		PolicyArn: aws.String(arn),
	})
	if err != nil {
		return err
	}
	return nil
}
