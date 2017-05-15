package main

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"log"
	"github.com/hashicorp/terraform/helper/resource"
	"time"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"fmt"
	"github.com/jckuester/awsweeper/schema"
)

func resourceAwsVpcDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	vpcID := d.Id()
	DeleteVpcOpts := &ec2.DeleteVpcInput{
		VpcId: &vpcID,
	}
	log.Printf("[INFO] Deleting VPC: %s", d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteVpc(DeleteVpcOpts)
		if err == nil {
			return nil
		}

		ec2err, ok := err.(awserr.Error)
		if !ok {
			return resource.NonRetryableError(err)
		}

		switch ec2err.Code() {
		case "InvalidVpcID.NotFound":
			return nil
		case "DependencyViolation":
			return resource.RetryableError(err)
		}

		return resource.NonRetryableError(fmt.Errorf("Error deleting VPC: %s", err))
	})
}
