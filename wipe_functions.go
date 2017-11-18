package main

import (
	"strings"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
)

func (c *WipeCommand) deleteGeneric(res Resources) {
	ids := []*string{}
	tags := []*map[string]string{}

	for i, r := range res.ids {
		if c.inCfg(res.ttype, r, res.tags[i]) {
			ids = append(ids, r)
			tags = append(tags, res.tags[i])
		}
	}
	c.wipe(Resources{ttype: res.ttype, ids: ids, tags: tags})
}

func (c *WipeCommand) deleteInstances(res Resources) {
	ids := []*string{}
	tags := []*map[string]string{}

	for _, r := range res.raw.(*ec2.DescribeInstancesOutput).Reservations {
		for _, in := range r.Instances {
			if *in.State.Name != "terminated" {
				m := &map[string]string{}
				for _, t := range in.Tags {
					(*m)[*t.Key] = *t.Value
				}

				if c.inCfg(res.ttype, in.InstanceId, m) {
					ids = append(ids, in.InstanceId)
					tags = append(tags, m)
				}
			}
		}
	}
	c.wipe(Resources{ttype: res.ttype, ids: ids, tags: tags})
}

func (c *WipeCommand) deleteInternetGateways(res Resources) {
	ids := []*string{}
	attrs := []*map[string]string{}
	tags := []*map[string]string{}

	for _, r := range res.raw.(*ec2.DescribeInternetGatewaysOutput).InternetGateways {
		m := &map[string]string{}
		for _, t := range r.Tags {
			(*m)[*t.Key] = *t.Value
		}

		if c.inCfg(res.ttype, r.InternetGatewayId, m) {
			ids = append(ids, r.InternetGatewayId)
			attrs = append(attrs, &map[string]string{
				"vpc_id": *r.Attachments[0].VpcId,
			})
			tags = append(tags, m)
		}
	}
	c.wipe(Resources{ttype: res.ttype, ids: ids, attrs: attrs, tags: tags})
}

func (c *WipeCommand) deleteNatGateways(res Resources) {
	ids := []*string{}

	for _, r := range res.raw.(*ec2.DescribeNatGatewaysOutput).NatGateways {
		if c.inCfg(res.ttype, r.NatGatewayId) {
			if *r.State == "available" {
				ids = append(ids, r.NatGatewayId)
			}
		}
	}
	c.wipe(Resources{ttype: res.ttype, ids: ids})
}

func (c *WipeCommand) deleteRoute53Record(res Resources) {
	ids := []*string{}

	// HostedZoneId is a required field for input
	for _, r := range res.raw.(*route53.ListResourceRecordSetsOutput).ResourceRecordSets {
		for _, rr := range r.ResourceRecords {
			if c.inCfg(res.ttype, rr.Value) {
				ids = append(ids, rr.Value)
			}
		}
	}
	c.wipe(Resources{ttype: res.ttype, ids: ids})
}

func (c *WipeCommand) deleteRoute53Zone(res Resources) {
	hzIds := []*string{}
	rsIds := []*string{}
	rsAttrs := []*map[string]string{}
	hzAttrs := []*map[string]string{}

	for _, hz := range res.raw.(*route53.ListHostedZonesOutput).HostedZones {
		if c.inCfg(res.ttype, hz.Id) {
			res, err := c.client.r53conn.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{
				HostedZoneId: hz.Id,
			})
			check(err)

			for _, rs := range res.ResourceRecordSets {
				rsIds = append(rsIds, rs.Name)
				rsAttrs = append(rsAttrs, &map[string]string{
					"zone_id": *hz.Id,
					"name":    *rs.Name,
					"type":    *rs.Type,
				})
			}
			hzIds = append(hzIds, hz.Id)
			hzAttrs = append(hzAttrs, &map[string]string{
				"force_destroy": "true",
				"name":          *hz.Name,
			})
		}
	}
	c.wipe(Resources{ttype: res.ttype, ids: hzIds, attrs: hzAttrs})
}

func (c *WipeCommand) deleteEfsFileSystem(res Resources) {
	fsIds := []*string{}
	mtIds := []*string{}

	for _, r := range res.raw.(*efs.DescribeFileSystemsOutput).FileSystems {
		if c.inCfg(res.ttype, r.Name) {
			res, err := c.client.efsconn.DescribeMountTargets(&efs.DescribeMountTargetsInput{
				FileSystemId: r.FileSystemId,
			})

			if err == nil {
				for _, r := range res.MountTargets {
					mtIds = append(mtIds, r.MountTargetId)
				}
			}

			fsIds = append(fsIds, r.FileSystemId)
		}
	}
	c.wipe(Resources{ttype: "aws_efs_mount_target", ids: mtIds})
	c.wipe(Resources{ttype: res.ttype, ids: fsIds})
}

func (c *WipeCommand) deleteIamUser(res Resources) {
	ids := []*string{}
	pIds := []*string{}
	upIds := []*string{}
	attrs := []*map[string]string{}
	pAttrs := []*map[string]string{}

	for _, u := range res.raw.(*iam.ListUsersOutput).Users {
		if c.inCfg(res.ttype, u.UserName) {

			// list inline policies, delete with "aws_iam_user_policy" delete routine
			ups, err := c.client.iamconn.ListUserPolicies(&iam.ListUserPoliciesInput{
				UserName: u.UserName,
			})
			if err == nil {
				for _, up := range ups.PolicyNames {
					upIds = append(upIds, aws.String(*u.UserName + ":" + *up))
				}
			}

			// Lists all managed policies that are attached  to user (inline and others)
			upols, err := c.client.iamconn.ListAttachedUserPolicies(&iam.ListAttachedUserPoliciesInput{
				// required
				UserName: u.UserName,
			})
			if err == nil {
				for _, upol := range upols.AttachedPolicies {
					pIds = append(pIds, upol.PolicyArn)
					pAttrs = append(pAttrs, &map[string]string{
						"user":       *u.UserName,
						"policy_arn": *upol.PolicyArn,
					})
				}
			}

			ids = append(ids, u.UserName)
			attrs = append(attrs, &map[string]string{
				"force_destroy": "true",
			})
		}
	}
	// aws_iam_user_policy to delete inline policies
	c.wipe(Resources{ttype: "aws_iam_user_policy", ids: upIds})
	c.wipe(Resources{ttype: "aws_iam_user_policy_attachment", ids: pIds, attrs: pAttrs})
	c.wipe(Resources{ttype: res.ttype, ids: ids, attrs: attrs})
}

func (c *WipeCommand) deleteIamPolicy(res Resources) {
	ids := []*string{}
	eIds := []*string{}
	attributes := []*map[string]string{}

	for _, pol := range res.raw.(*iam.ListPoliciesOutput).Policies {
		if c.inCfg(res.ttype, pol.Arn) {
			es, err := c.client.iamconn.ListEntitiesForPolicy(&iam.ListEntitiesForPolicyInput{
				PolicyArn: pol.Arn,
			})
			check(err)

			roles := []string{}
			users := []string{}
			groups := []string{}

			for _, u := range es.PolicyUsers {
				users = append(users, *u.UserName)
			}
			for _, g := range es.PolicyGroups {
				groups = append(groups, *g.GroupName)
			}
			for _, r := range es.PolicyRoles {
				roles = append(roles, *r.RoleName)
			}

			eIds = append(eIds, pol.Arn)
			attributes = append(attributes, &map[string]string{
				"policy_arn": *pol.Arn,
				"name":       *pol.PolicyName,
				"users":      strings.Join(users, "."),
				"roles":      strings.Join(roles, "."),
				"groups":     strings.Join(groups, "."),
			})
			ids = append(ids, pol.Arn)
		}
	}
	// policy attachments are not resources
	// what happens here, is that policy is detached from groups, users and roles
	c.wipe(Resources{ttype: "aws_iam_policy_attachment", ids: eIds, attrs: attributes})
	c.wipe(Resources{ttype: res.ttype, ids: ids})
}

func (c *WipeCommand) deleteIamRole(res Resources) {
	ids := []*string{}
	rpolIds := []*string{}
	rpolAttributes := []*map[string]string{}
	pIds := []*string{}

	for _, role := range res.raw.(*iam.ListRolesOutput).Roles {
		if c.inCfg(res.ttype, role.RoleName) {
			rpols, err := c.client.iamconn.ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{
				RoleName: role.RoleName,
			})
			check(err)

			for _, rpol := range rpols.AttachedPolicies {
				rpolIds = append(rpolIds, rpol.PolicyArn)
				rpolAttributes = append(rpolAttributes, &map[string]string{
					"role":       *role.RoleName,
					"policy_arn": *rpol.PolicyArn,
				})
			}

			rps, err := c.client.iamconn.ListRolePolicies(&iam.ListRolePoliciesInput{
				RoleName: role.RoleName,
			})
			check(err)

			for _, rp := range rps.PolicyNames {
				bla := *role.RoleName + ":" + *rp
				pIds = append(pIds, &bla)
			}

			//ips, err := c.client.iamconn.ListInstanceProfilesForRole(&iam.ListInstanceProfilesForRoleInput{
			//	RoleName: role.RoleName,
			//})
			//check(err)
			//
			//for _, ip := range ips.InstanceProfiles {
			//	fmt.Println(*ip.InstanceProfileName)
			//}

			ids = append(ids, role.RoleName)
		}
	}

	// aws_iam_policy_attachment could be used to detach a policy from users, groups and roles
	c.wipe(Resources{ttype: "aws_iam_role_policy_attachment", ids: rpolIds, attrs: rpolAttributes})
	c.wipe(Resources{ttype: "aws_iam_role_policy", ids: pIds})
	c.wipe(Resources{ttype: res.ttype, ids: ids})
}

func (c *WipeCommand) deleteInstanceProfiles(res Resources) {
	ids := []*string{}
	attributes := []*map[string]string{}

	for _, r := range res.raw.(*iam.ListInstanceProfilesOutput).InstanceProfiles {
		if c.inCfg(res.ttype, r.InstanceProfileName) {
			ids = append(ids, r.InstanceProfileName)

			roles := []string{}
			for _, role := range r.Roles {
				roles = append(roles, *role.RoleName)
			}

			attributes = append(attributes, &map[string]string{
				"roles": strings.Join(roles, "."),
			})
		}
	}
	c.wipe(Resources{ttype: res.ttype, ids: ids})
}

func (c *WipeCommand) deleteKmsKeys(res Resources) {
	ids := []*string{}
	attributes := []*map[string]string{}

	for _, r := range res.raw.(*kms.ListKeysOutput).Keys {
		if c.inCfg(res.ttype, r.KeyArn) {
			req, res := c.client.kmsconn.DescribeKeyRequest(&kms.DescribeKeyInput{
				KeyId: r.KeyId,
			})
			err := req.Send();
			if err == nil {
				if *res.KeyMetadata.KeyState != "PendingDeletion" {
					attributes = append(attributes, &map[string]string{
						"key_id": *r.KeyId,
					})
					ids = append(ids, r.KeyArn)
				}
			}
		}
	}
	c.wipe(Resources{ttype: res.ttype, ids: ids})
}

func (c *WipeCommand) deleteAmis(res Resources) {
	ids := []*string{}
	tags := []*map[string]string{}

	accountId := *c.getAccountId()

	for _, r := range res.raw.(*ec2.DescribeImagesOutput).Images {
		m := &map[string]string{}
		for _, t := range r.Tags {
			(*m)[*t.Key] = *t.Value
		}

		if accountId == *r.OwnerId && c.inCfg(res.ttype, r.ImageId, m) {
			ids = append(ids, r.ImageId)
			tags = append(tags, m)
		}
	}
	c.wipe(Resources{ttype: res.ttype, ids: ids, tags: tags})
}

func (c *WipeCommand) deleteSnapshots(res Resources) {
	ids := []*string{}
	tags := []*map[string]string{}

	accountId := *c.getAccountId()

	for _, r := range res.raw.(*ec2.DescribeSnapshotsOutput).Snapshots {
		m := &map[string]string{}
		for _, t := range r.Tags {
			(*m)[*t.Key] = *t.Value
		}

		if accountId == *r.OwnerId && c.inCfg(res.ttype, r.SnapshotId, m) {
			ids = append(ids, r.SnapshotId)
			tags = append(tags, m)
		}
	}
	c.wipe(Resources{ttype: res.ttype, ids: ids, tags: tags})
}

func (c *WipeCommand) getAccountId() *string {
	res, err := c.client.stsconn.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	check(err)
	return res.Account
}
