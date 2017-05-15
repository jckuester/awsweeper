package main

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"log"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/jckuester/awsweeper/schema"
)

func resourceAwsVPCEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	input := &ec2.DeleteVpcEndpointsInput{
		VpcEndpointIds: []*string{aws.String(d.Id())},
	}

	log.Printf("[DEBUG] Deleting VPC Endpoint: %#v", input)
	_, err := conn.DeleteVpcEndpoints(input)

	if err != nil {
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return fmt.Errorf("Error deleting VPC Endpoint: %s", err.Error())
		}

		if ec2err.Code() == "InvalidVpcEndpointId.NotFound" {
			log.Printf("[DEBUG] VPC Endpoint %q is already gone", d.Id())
		} else {
			return fmt.Errorf("Error deleting VPC Endpoint: %s", err.Error())
		}
	}

	log.Printf("[DEBUG] VPC Endpoint %q deleted", d.Id())
	d.SetId("")

	return nil
}
