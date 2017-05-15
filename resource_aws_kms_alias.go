package main

import (
	"github.com/aws/aws-sdk-go/service/kms"
	"log"
	"github.com/jckuester/awsweeper/schema"
	"github.com/aws/aws-sdk-go/aws"
)

func resourceAwsKmsAliasDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn

	req := &kms.DeleteAliasInput{
		AliasName: aws.String(d.Id()),
	}
	_, err := conn.DeleteAlias(req)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] KMS Alias: (%s) deleted.", d.Id())
	d.SetId("")
	return nil
}
