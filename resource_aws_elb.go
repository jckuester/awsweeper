package main

import (
	"log"
	"github.com/aws/aws-sdk-go/service/elb"
	"fmt"
	"github.com/jckuester/awsweeper/schema"
	"github.com/aws/aws-sdk-go/aws"
)

func resourceAwsElbDelete(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbconn

	log.Printf("[INFO] Deleting ELB: %s", d.Id())

	// Destroy the load balancer
	deleteElbOpts := elb.DeleteLoadBalancerInput{
		LoadBalancerName: aws.String(d.Id()),
	}
	if _, err := elbconn.DeleteLoadBalancer(&deleteElbOpts); err != nil {
		return fmt.Errorf("Error deleting ELB: %s", err)
	}

	return nil
}
