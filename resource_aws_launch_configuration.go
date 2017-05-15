package main

import (
	"log"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/jckuester/awsweeper/schema"
	"github.com/aws/aws-sdk-go/aws"
)

func resourceAwsLaunchConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	autoscalingconn := meta.(*AWSClient).autoscalingconn

	log.Printf("[DEBUG] Launch Configuration destroy: %v", d.Id())
	_, err := autoscalingconn.DeleteLaunchConfiguration(
		&autoscaling.DeleteLaunchConfigurationInput{
			LaunchConfigurationName: aws.String(d.Id()),
		})
	if err != nil {
		autoscalingerr, ok := err.(awserr.Error)
		if ok && (autoscalingerr.Code() == "InvalidConfiguration.NotFound" || autoscalingerr.Code() == "ValidationError") {
			log.Printf("[DEBUG] Launch configuration (%s) not found", d.Id())
			return nil
		}

		return err
	}

	return nil
}
