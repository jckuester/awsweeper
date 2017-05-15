package main

import (
	"log"
	"github.com/hashicorp/terraform/helper/resource"
	"time"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"fmt"
	"github.com/jckuester/awsweeper/schema"
	"github.com/aws/aws-sdk-go/aws"
)

func resourceAwsInternetGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	// Detach if it is attached
	if err := resourceAwsInternetGatewayDetach(d, meta); err != nil {
		return err
	}

	log.Printf("[INFO] Deleting Internet Gateway: %s", d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteInternetGateway(&ec2.DeleteInternetGatewayInput{
			InternetGatewayId: aws.String(d.Id()),
		})
		if err == nil {
			return nil
		}

		ec2err, ok := err.(awserr.Error)
		if !ok {
			return resource.RetryableError(err)
		}

		switch ec2err.Code() {
		case "InvalidInternetGatewayID.NotFound":
			return nil
		case "DependencyViolation":
			return resource.RetryableError(err) // retry
		}

		return resource.NonRetryableError(err)
	})
}


func resourceAwsInternetGatewayDetach(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	// Get the old VPC ID to detach from
	vpcID, _ := d.GetChange("vpc_id")

	if vpcID.(string) == "" {
		log.Printf(
			"[DEBUG] Not detaching Internet Gateway '%s' as no VPC ID is set",
			d.Id())
		return nil
	}

	log.Printf(
		"[INFO] Detaching Internet Gateway '%s' from VPC '%s'",
		d.Id(),
		vpcID.(string))

	// Wait for it to be fully detached before continuing
	log.Printf("[DEBUG] Waiting for internet gateway (%s) to detach", d.Id())
	stateConf := &resource.StateChangeConf{
		Pending:        []string{"detaching"},
		Target:         []string{"detached"},
		Refresh:        detachIGStateRefreshFunc(conn, d.Id(), vpcID.(string)),
		Timeout:        15 * time.Minute,
		Delay:          10 * time.Second,
		NotFoundChecks: 30,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf(
			"Error waiting for internet gateway (%s) to detach: %s",
			d.Id(), err)
	}

	return nil
}

// InstanceStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch
// an EC2 instance.
func detachIGStateRefreshFunc(conn *ec2.EC2, gatewayID, vpcID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		_, err := conn.DetachInternetGateway(&ec2.DetachInternetGatewayInput{
			InternetGatewayId: aws.String(gatewayID),
			VpcId:             aws.String(vpcID),
		})
		if err != nil {
			if ec2err, ok := err.(awserr.Error); ok {
				switch ec2err.Code() {
				case "InvalidInternetGatewayID.NotFound":
					log.Printf("[TRACE] Error detaching Internet Gateway '%s' from VPC '%s': %s", gatewayID, vpcID, err)
					return nil, "Not Found", nil

				case "Gateway.NotAttached":
					return "detached", "detached", nil

				case "DependencyViolation":
					return nil, "detaching", nil
				}
			}
		}

		// DetachInternetGateway only returns an error, so if it's nil, assume we're
		// detached
		return "detached", "detached", nil
	}
}
