package main

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"log"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/hashicorp/terraform/helper/resource"
	"time"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/jckuester/awsweeper/schema"
)

func resourceAwsNatGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	deleteOpts := &ec2.DeleteNatGatewayInput{
		NatGatewayId: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting NAT Gateway: %s", d.Id())

	_, err := conn.DeleteNatGateway(deleteOpts)
	if err != nil {
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}

		if ec2err.Code() == "NatGatewayNotFound" {
			return nil
		}

		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"deleting"},
		Target:     []string{"deleted"},
		Refresh:    NGStateRefreshFunc(conn, d.Id()),
		Timeout:    30 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	_, stateErr := stateConf.WaitForState()
	if stateErr != nil {
		return fmt.Errorf("Error waiting for NAT Gateway (%s) to delete: %s", d.Id(), err)
	}

	return nil
}

// NGStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch
// a NAT Gateway.
func NGStateRefreshFunc(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		opts := &ec2.DescribeNatGatewaysInput{
			NatGatewayIds: []*string{aws.String(id)},
		}
		resp, err := conn.DescribeNatGateways(opts)
		if err != nil {
			if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "NatGatewayNotFound" {
				resp = nil
			} else {
				log.Printf("Error on NGStateRefresh: %s", err)
				return nil, "", err
			}
		}

		if resp == nil {
			// Sometimes AWS just has consistency issues and doesn't see
			// our instance yet. Return an empty state.
			return nil, "", nil
		}

		ng := resp.NatGateways[0]
		return ng, *ng.State, nil
	}
}
