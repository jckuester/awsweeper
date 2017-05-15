package main

import (
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"fmt"
	"github.com/jckuester/awsweeper/schema"
	"github.com/aws/aws-sdk-go/aws"
)

func resourceAwsIamPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	if err := iamPolicyDeleteNondefaultVersions(d.Id(), iamconn); err != nil {
		return err
	}

	request := &iam.DeletePolicyInput{
		PolicyArn: aws.String(d.Id()),
	}

	_, err := iamconn.DeletePolicy(request)
	if err != nil {
		if iamerr, ok := err.(awserr.Error); ok && iamerr.Code() == "NoSuchEntity" {
			return nil
		}
		return fmt.Errorf("Error reading IAM policy %s: %#v", d.Id(), err)
	}
	return nil
}

func iamPolicyDeleteNondefaultVersions(arn string, iamconn *iam.IAM) error {
	versions, err := iamPolicyListVersions(arn, iamconn)
	if err != nil {
		return err
	}

	for _, version := range versions {
		if *version.IsDefaultVersion {
			continue
		}
		if err := iamPolicyDeleteVersion(arn, *version.VersionId, iamconn); err != nil {
			return err
		}
	}

	return nil
}

func iamPolicyDeleteVersion(arn, versionID string, iamconn *iam.IAM) error {
	request := &iam.DeletePolicyVersionInput{
		PolicyArn: aws.String(arn),
		VersionId: aws.String(versionID),
	}

	_, err := iamconn.DeletePolicyVersion(request)
	if err != nil {
		return fmt.Errorf("Error deleting version %s from IAM policy %s: %s", versionID, arn, err)
	}
	return nil
}

func iamPolicyListVersions(arn string, iamconn *iam.IAM) ([]*iam.PolicyVersion, error) {
	request := &iam.ListPolicyVersionsInput{
		PolicyArn: aws.String(arn),
	}

	response, err := iamconn.ListPolicyVersions(request)
	if err != nil {
		return nil, fmt.Errorf("Error listing versions for IAM policy %s: %s", arn, err)
	}
	return response.Versions, nil
}
