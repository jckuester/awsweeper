package main

import (
	"log"
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"time"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/jckuester/awsweeper/schema"
	"strconv"
)

func resourceAwsSecurityGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[DEBUG] Security Group destroy: %v", d.Id())

	if err := deleteLingeringLambdaENIs(conn, d); err != nil {
		return fmt.Errorf("Failed to delete Lambda ENIs: %s", err)
	}

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{
			GroupId: aws.String(d.Id()),
		})
		if err != nil {
			ec2err, ok := err.(awserr.Error)
			if !ok {
				return resource.RetryableError(err)
			}

			switch ec2err.Code() {
			case "InvalidGroup.NotFound":
				return nil
			case "DependencyViolation":
				// If it is a dependency violation, we want to retry
				return resource.RetryableError(err)
			default:
				// Any other error, we want to quit the retry loop immediately
				return resource.NonRetryableError(err)
			}
		}

		return nil
	})
}

// The AWS Lambda service creates ENIs behind the scenes and keeps these around for a while
// which would prevent SGs attached to such ENIs from being destroyed
func deleteLingeringLambdaENIs(conn *ec2.EC2, d *schema.ResourceData) error {
	// Here we carefully find the offenders
	params := &ec2.DescribeNetworkInterfacesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("group-id"),
				Values: []*string{aws.String(d.Id())},
			},
			{
				Name:   aws.String("description"),
				Values: []*string{aws.String("AWS Lambda VPC ENI: *")},
			},
			{
				Name:   aws.String("requester-id"),
				Values: []*string{aws.String("*:awslambda_*")},
			},
		},
	}
	networkInterfaceResp, err := conn.DescribeNetworkInterfaces(params)
	if err != nil {
		return err
	}

	// Then we detach and finally delete those
	v := networkInterfaceResp.NetworkInterfaces
	for _, eni := range v {
		if eni.Attachment != nil {
			detachNetworkInterfaceParams := &ec2.DetachNetworkInterfaceInput{
				AttachmentId: eni.Attachment.AttachmentId,
			}
			_, detachNetworkInterfaceErr := conn.DetachNetworkInterface(detachNetworkInterfaceParams)

			if detachNetworkInterfaceErr != nil {
				return detachNetworkInterfaceErr
			}

			log.Printf("[DEBUG] Waiting for ENI (%s) to become detached", *eni.NetworkInterfaceId)
			stateConf := &resource.StateChangeConf{
				Pending: []string{"true"},
				Target:  []string{"false"},
				Refresh: networkInterfaceAttachedRefreshFunc(conn, *eni.NetworkInterfaceId),
				Timeout: 10 * time.Minute,
			}
			if _, err := stateConf.WaitForState(); err != nil {
				return fmt.Errorf(
					"Error waiting for ENI (%s) to become detached: %s", *eni.NetworkInterfaceId, err)
			}
		}

		deleteNetworkInterfaceParams := &ec2.DeleteNetworkInterfaceInput{
			NetworkInterfaceId: eni.NetworkInterfaceId,
		}
		_, deleteNetworkInterfaceErr := conn.DeleteNetworkInterface(deleteNetworkInterfaceParams)

		if deleteNetworkInterfaceErr != nil {
			return deleteNetworkInterfaceErr
		}
	}

	return nil
}


func networkInterfaceAttachedRefreshFunc(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		describe_network_interfaces_request := &ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []*string{aws.String(id)},
		}
		describeResp, err := conn.DescribeNetworkInterfaces(describe_network_interfaces_request)

		if err != nil {
			log.Printf("[ERROR] Could not find network interface %s. %s", id, err)
			return nil, "", err
		}

		eni := describeResp.NetworkInterfaces[0]
		hasAttachment := strconv.FormatBool(eni.Attachment != nil)
		log.Printf("[DEBUG] ENI %s has attachment state %s", id, hasAttachment)
		return eni, hasAttachment, nil
	}
}
