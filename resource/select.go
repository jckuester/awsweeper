package resource

import (
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/sts"
)

func filterGeneric(res Resources, f Filter, c *AWSClient) []Resources {
	ids := []*string{}
	tags := []*map[string]string{}

	for i, r := range res.Ids {
		if f.Matches(res.Type, *r, *res.Tags[i]) {
			ids = append(ids, r)
			tags = append(tags, res.Tags[i])
		}
	}

	return []Resources{{Type: res.Type, Ids: ids, Tags: tags}}
}

func filterInstances(res Resources, f Filter, c *AWSClient) []Resources {
	ids := []*string{}
	tags := []*map[string]string{}

	for _, r := range res.Raw.(*ec2.DescribeInstancesOutput).Reservations {
		for _, in := range r.Instances {
			if *in.State.Name != "terminated" {
				m := &map[string]string{}
				for _, t := range in.Tags {
					(*m)[*t.Key] = *t.Value
				}

				if f.Matches(res.Type, *in.InstanceId, *m) {
					ids = append(ids, in.InstanceId)
					tags = append(tags, m)
				}
			}
		}
	}

	return []Resources{{Type: res.Type, Ids: ids, Tags: tags}}
}

func filterInternetGateways(res Resources, f Filter, c *AWSClient) []Resources {
	ids := []*string{}
	attrs := []*map[string]string{}
	tags := []*map[string]string{}

	for _, r := range res.Raw.(*ec2.DescribeInternetGatewaysOutput).InternetGateways {
		m := &map[string]string{}
		for _, t := range r.Tags {
			(*m)[*t.Key] = *t.Value
		}

		if f.Matches(res.Type, *r.InternetGatewayId, *m) {
			ids = append(ids, r.InternetGatewayId)
			attrs = append(attrs, &map[string]string{
				"vpc_id": *r.Attachments[0].VpcId,
			})
			tags = append(tags, m)
		}
	}

	return []Resources{{Type: res.Type, Ids: ids, Attrs: attrs, Tags: tags}}
}

func filterNatGateways(res Resources, f Filter, c *AWSClient) []Resources {
	ids := []*string{}

	for _, r := range res.Raw.(*ec2.DescribeNatGatewaysOutput).NatGateways {
		if f.Matches(res.Type, *r.NatGatewayId) {
			if *r.State == "available" {
				ids = append(ids, r.NatGatewayId)
			}
		}
	}
	return []Resources{{Type: res.Type, Ids: ids}}
}

func filterRoute53Record(res Resources, f Filter, c *AWSClient) []Resources {
	ids := []*string{}

	// HostedZoneId is a required field for input
	for _, r := range res.Raw.(*route53.ListResourceRecordSetsOutput).ResourceRecordSets {
		for _, rr := range r.ResourceRecords {
			if f.Matches(res.Type, *rr.Value) {
				ids = append(ids, rr.Value)
			}
		}
	}
	return []Resources{{Type: res.Type, Ids: ids}}
}

func filterRoute53Zone(res Resources, f Filter, c *AWSClient) []Resources {
	hzIds := []*string{}
	rsIds := []*string{}
	rsAttrs := []*map[string]string{}
	hzAttrs := []*map[string]string{}

	for _, hz := range res.Raw.(*route53.ListHostedZonesOutput).HostedZones {
		if f.Matches(res.Type, *hz.Id) {
			res, err := c.R53conn.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{
				HostedZoneId: hz.Id,
			})
			if err != nil {
				log.Fatal(err)
			}

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
	return []Resources{{Type: res.Type, Ids: hzIds, Attrs: hzAttrs}}
}

func filterEfsFileSystem(res Resources, f Filter, c *AWSClient) []Resources {
	fsIds := []*string{}
	mtIds := []*string{}

	for _, r := range res.Raw.(*efs.DescribeFileSystemsOutput).FileSystems {
		if f.Matches(res.Type, *r.Name) {
			res, err := c.EFSconn.DescribeMountTargets(&efs.DescribeMountTargetsInput{
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

	return []Resources{
		{Type: "aws_efs_mount_target", Ids: mtIds},
		{Type: res.Type, Ids: fsIds},
	}
}

func filterIamUser(res Resources, f Filter, c *AWSClient) []Resources {
	ids := []*string{}
	pIds := []*string{}
	upIds := []*string{}
	attrs := []*map[string]string{}
	pAttrs := []*map[string]string{}

	for _, u := range res.Raw.(*iam.ListUsersOutput).Users {
		if f.Matches(res.Type, *u.UserName) {
			// list inline policies, delete with "aws_iam_user_policy" delete routine
			ups, err := c.IAMconn.ListUserPolicies(&iam.ListUserPoliciesInput{
				UserName: u.UserName,
			})
			if err == nil {
				for _, up := range ups.PolicyNames {
					upIds = append(upIds, aws.String(*u.UserName+":"+*up))
				}
			}

			// Lists all managed policies that are attached  to user (inline and others)
			upols, err := c.IAMconn.ListAttachedUserPolicies(&iam.ListAttachedUserPoliciesInput{
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
	return []Resources{
		{Type: "aws_iam_user_policy", Ids: upIds},
		{Type: "aws_iam_user_policy_attachment", Ids: pIds, Attrs: pAttrs},
		{Type: res.Type, Ids: ids, Attrs: attrs},
	}
}

func filterIamPolicy(res Resources, f Filter, c *AWSClient) []Resources {
	ids := []*string{}
	eIds := []*string{}
	attributes := []*map[string]string{}

	for _, pol := range res.Raw.(*iam.ListPoliciesOutput).Policies {
		if f.Matches(res.Type, *pol.Arn) {
			es, err := c.IAMconn.ListEntitiesForPolicy(&iam.ListEntitiesForPolicyInput{
				PolicyArn: pol.Arn,
			})
			if err != nil {
				log.Fatal(err)
			}

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
	return []Resources{
		{Type: "aws_iam_policy_attachment", Ids: eIds, Attrs: attributes},
		{Type: res.Type, Ids: ids},
	}
}

func filterIamRole(res Resources, f Filter, c *AWSClient) []Resources {
	ids := []*string{}
	rpolIds := []*string{}
	rpolAttributes := []*map[string]string{}
	pIds := []*string{}

	for _, role := range res.Raw.(*iam.ListRolesOutput).Roles {
		if f.Matches(res.Type, *role.RoleName) {
			rpols, err := c.IAMconn.ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{
				RoleName: role.RoleName,
			})
			if err != nil {
				log.Fatal(err)
			}

			for _, rpol := range rpols.AttachedPolicies {
				rpolIds = append(rpolIds, rpol.PolicyArn)
				rpolAttributes = append(rpolAttributes, &map[string]string{
					"role":       *role.RoleName,
					"policy_arn": *rpol.PolicyArn,
				})
			}

			rps, err := c.IAMconn.ListRolePolicies(&iam.ListRolePoliciesInput{
				RoleName: role.RoleName,
			})
			if err != nil {
				log.Fatal(err)
			}

			for _, rp := range rps.PolicyNames {
				bla := *role.RoleName + ":" + *rp
				pIds = append(pIds, &bla)
			}

			//ips, err := c.iamconn.ListInstanceProfilesForRole(&iam.ListInstanceProfilesForRoleInput{
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
	return []Resources{
		{Type: "aws_iam_role_policy_attachment", Ids: rpolIds, Attrs: rpolAttributes},
		{Type: "aws_iam_role_policy", Ids: pIds},
		{Type: res.Type, Ids: ids},
	}
}

func filterInstanceProfiles(res Resources, f Filter, c *AWSClient) []Resources {
	ids := []*string{}
	attributes := []*map[string]string{}

	for _, r := range res.Raw.(*iam.ListInstanceProfilesOutput).InstanceProfiles {
		if f.Matches(res.Type, *r.InstanceProfileName) {
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
	return []Resources{{Type: res.Type, Ids: ids}}
}

func filterKmsKeys(res Resources, f Filter, c *AWSClient) []Resources {
	ids := []*string{}
	attributes := []*map[string]string{}

	for _, r := range res.Raw.(*kms.ListKeysOutput).Keys {
		if f.Matches(res.Type, *r.KeyArn) {
			req, res := c.KMSconn.DescribeKeyRequest(&kms.DescribeKeyInput{
				KeyId: r.KeyId,
			})
			err := req.Send()
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
	return []Resources{{Type: res.Type, Ids: ids}}
}

func filterAmis(res Resources, f Filter, c *AWSClient) []Resources {
	ids := []*string{}
	tags := []*map[string]string{}

	for _, r := range res.Raw.(*ec2.DescribeImagesOutput).Images {
		m := &map[string]string{}
		for _, t := range r.Tags {
			(*m)[*t.Key] = *t.Value
		}

		if accountId(c) == *r.OwnerId && (f.Matches(res.Type, *r.ImageId, *m)) {
			ids = append(ids, r.ImageId)
			tags = append(tags, m)
		}
	}
	return []Resources{{Type: res.Type, Ids: ids, Tags: tags}}
}

func filterSnapshots(res Resources, f Filter, c *AWSClient) []Resources {
	ids := []*string{}
	tags := []*map[string]string{}

	for _, r := range res.Raw.(*ec2.DescribeSnapshotsOutput).Snapshots {
		m := &map[string]string{}
		for _, t := range r.Tags {
			(*m)[*t.Key] = *t.Value
		}

		if accountId(c) == *r.OwnerId && (f.Matches(res.Type, *r.SnapshotId, *m)) {
			ids = append(ids, r.SnapshotId)
			tags = append(tags, m)
		}
	}
	return []Resources{{Type: res.Type, Ids: ids, Tags: tags}}
}

func accountId(c *AWSClient) string {
	res, err := c.STSconn.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		log.Fatal(err)
	}

	return *res.Account
}
