package main

import (
	"github.com/aws/aws-sdk-go/service/iam"
	"fmt"
	"log"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/jckuester/awsweeper/schema"
	"github.com/aws/aws-sdk-go/aws"
)

func resourceAwsIamUserDelete(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	// IAM Users must be removed from all groups before they can be deleted
	var groups []string
	listGroups := &iam.ListGroupsForUserInput{
		UserName: aws.String(d.Id()),
	}
	pageOfGroups := func(page *iam.ListGroupsForUserOutput, lastPage bool) (shouldContinue bool) {
		for _, g := range page.Groups {
			groups = append(groups, *g.GroupName)
		}
		return !lastPage
	}
	err := iamconn.ListGroupsForUserPages(listGroups, pageOfGroups)
	if err != nil {
		return fmt.Errorf("Error removing user %q from all groups: %s", d.Id(), err)
	}
	for _, g := range groups {
		// use iam group membership func to remove user from all groups
		log.Printf("[DEBUG] Removing IAM User %s from IAM Group %s", d.Id(), g)
		if err := removeUsersFromGroup(iamconn, []*string{aws.String(d.Id())}, g); err != nil {
			return err
		}
	}

	// All access keys, MFA devices and login profile for the user must be removed
	if d.Get("force_destroy").(bool) {
		var accessKeys []string
		listAccessKeys := &iam.ListAccessKeysInput{
			UserName: aws.String(d.Id()),
		}
		pageOfAccessKeys := func(page *iam.ListAccessKeysOutput, lastPage bool) (shouldContinue bool) {
			for _, k := range page.AccessKeyMetadata {
				accessKeys = append(accessKeys, *k.AccessKeyId)
			}
			return !lastPage
		}
		err = iamconn.ListAccessKeysPages(listAccessKeys, pageOfAccessKeys)
		if err != nil {
			return fmt.Errorf("Error removing access keys of user %s: %s", d.Id(), err)
		}
		for _, k := range accessKeys {
			_, err := iamconn.DeleteAccessKey(&iam.DeleteAccessKeyInput{
				UserName:    aws.String(d.Id()),
				AccessKeyId: aws.String(k),
			})
			if err != nil {
				return fmt.Errorf("Error deleting access key %s: %s", k, err)
			}
		}

		var MFADevices []string
		listMFADevices := &iam.ListMFADevicesInput{
			UserName: aws.String(d.Id()),
		}
		pageOfMFADevices := func(page *iam.ListMFADevicesOutput, lastPage bool) (shouldContinue bool) {
			for _, m := range page.MFADevices {
				MFADevices = append(MFADevices, *m.SerialNumber)
			}
			return !lastPage
		}
		err = iamconn.ListMFADevicesPages(listMFADevices, pageOfMFADevices)
		if err != nil {
			return fmt.Errorf("Error removing MFA devices of user %s: %s", d.Id(), err)
		}
		for _, m := range MFADevices {
			_, err := iamconn.DeactivateMFADevice(&iam.DeactivateMFADeviceInput{
				UserName:     aws.String(d.Id()),
				SerialNumber: aws.String(m),
			})
			if err != nil {
				return fmt.Errorf("Error deactivating MFA device %s: %s", m, err)
			}
		}

		_, err = iamconn.DeleteLoginProfile(&iam.DeleteLoginProfileInput{
			UserName: aws.String(d.Id()),
		})
		if err != nil {
			if iamerr, ok := err.(awserr.Error); !ok || iamerr.Code() != "NoSuchEntity" {
				return fmt.Errorf("Error deleting Account Login Profile: %s", err)
			}
		}
	}

	request := &iam.DeleteUserInput{
		UserName: aws.String(d.Id()),
	}

	log.Println("[DEBUG] Delete IAM User request:", request)
	if _, err := iamconn.DeleteUser(request); err != nil {
		return fmt.Errorf("Error deleting IAM User %s: %s", d.Id(), err)
	}
	return nil
}

func removeUsersFromGroup(conn *iam.IAM, users []*string, group string) error {
	for _, u := range users {
		_, err := conn.RemoveUserFromGroup(&iam.RemoveUserFromGroupInput{
			UserName:  u,
			GroupName: aws.String(group),
		})

		if err != nil {
			if iamerr, ok := err.(awserr.Error); ok && iamerr.Code() == "NoSuchEntity" {
				return nil
			}
			return err
		}
	}
	return nil
}

