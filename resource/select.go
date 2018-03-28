package resource

import (
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/route53"
)

func filterGeneric(res Resources, raw interface{}, f Filter, c *AWSClient) []Resources {
	result := Resources{}

	for _, r := range res {
		if f.Matches(r.Type, r.Id, r.Tags) {
			result = append(result, r)
		}
	}
	return []Resources{result}
}

//func filterRoute53Record(res Resources, raw interface{}, f Filter, c *AWSClient) []Resources {
//	ids := []*string{}
//
//	// HostedZoneId is a required field for input
//	for _, r := range raw.(*route53.ListResourceRecordSetsOutput).ResourceRecordSets {
//		for _, rr := range r.ResourceRecords {
//			if f.Matches(res.Type, *rr.Value) {
//				ids = append(ids, rr.Value)
//			}
//		}
//	}
//	return []Resources{{Type: res.Type, Ids: ids}}
//}

func filterRoute53Zone(res Resources, raw interface{}, f Filter, c *AWSClient) []Resources {
	result := Resources{}

	//rsIds := []*string{}
	//rsAttrs := []*map[string]string{}

	for _, hz := range raw.(*route53.ListHostedZonesOutput).HostedZones {
		if f.Matches(res[0].Type, *hz.Id) {
			//res, err := c.R53conn.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{
			//	HostedZoneId: hz.Id,
			//})
			//if err != nil {
			//	log.Fatal(err)
			//}

			//for _, rs := range res.ResourceRecordSets {
			//	rsIds = append(rsIds, rs.Name)
			//	rsAttrs = append(rsAttrs, &map[string]string{
			//		"zone_id": *hz.Id,
			//		"name":    *rs.Name,
			//		"type":    *rs.Type,
			//	})
			//}

			result = append(result, &Resource{
				Type: res[0].Type,
				Id:   *hz.Id,
				Attrs: map[string]string{
					"force_destroy": "true",
					"name":          *hz.Name,
				},
			})
		}
	}
	return []Resources{result}
}

func filterEfsFileSystem(res Resources, raw interface{}, f Filter, c *AWSClient) []Resources {
	result := Resources{}
	resultMt := Resources{}

	for _, r := range res {
		if f.Matches(r.Type, *raw.(*efs.DescribeFileSystemsOutput).FileSystems[0].Name) {
			res, err := c.EFSconn.DescribeMountTargets(&efs.DescribeMountTargetsInput{
				FileSystemId: &r.Id,
			})

			if err == nil {
				for _, r := range res.MountTargets {
					resultMt = append(resultMt, &Resource{
						Type: "aws_efs_mount_target",
						Id:   *r.MountTargetId,
					})
				}
			}
			result = append(result, r)
		}
	}
	return []Resources{resultMt, result}
}

func filterIamUser(res Resources, raw interface{}, f Filter, c *AWSClient) []Resources {
	result := Resources{}
	resultAttPol := Resources{}
	resultUserPol := Resources{}

	for _, r := range res {
		if f.Matches(r.Type, r.Id) {
			// list inline policies, delete with "aws_iam_user_policy" delete routine
			ups, err := c.IAMconn.ListUserPolicies(&iam.ListUserPoliciesInput{
				UserName: &r.Id,
			})
			if err == nil {
				for _, up := range ups.PolicyNames {
					resultUserPol = append(resultUserPol, &Resource{
						Type: "aws_iam_user_policy",
						Id:   r.Id + ":" + *up,
					})
				}
			}

			// Lists all managed policies that are attached  to user (inline and others)
			upols, err := c.IAMconn.ListAttachedUserPolicies(&iam.ListAttachedUserPoliciesInput{
				// required
				UserName: &r.Id,
			})
			if err == nil {
				for _, upol := range upols.AttachedPolicies {
					resultAttPol = append(resultAttPol, &Resource{
						Type: "aws_iam_user_policy_attachment",
						Id:   *upol.PolicyArn,
						Attrs: map[string]string{
							"user":       r.Id,
							"policy_arn": *upol.PolicyArn,
						},
					})
				}
			}

			result = append(result, r)
			//attrs = append(attrs, &map[string]string{
			//	"force_destroy": "true",
			//})
		}
	}
	// aws_iam_user_policy to delete inline policies
	return []Resources{resultUserPol, resultAttPol, result}
}

func filterIamPolicy(res Resources, raw interface{}, f Filter, c *AWSClient) []Resources {
	result := Resources{}
	resultAtt := Resources{}

	for i, r := range res {
		if f.Matches(r.Type, r.Id) {
			es, err := c.IAMconn.ListEntitiesForPolicy(&iam.ListEntitiesForPolicyInput{
				PolicyArn: &r.Id,
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

			resultAtt = append(resultAtt, &Resource{
				Type: "aws_iam_policy_attachment",
				Id:   r.Id,
				Attrs: map[string]string{
					"policy_arn": r.Id,
					"name":       *raw.(*iam.ListPoliciesOutput).Policies[i].PolicyName,
					"users":      strings.Join(users, "."),
					"roles":      strings.Join(roles, "."),
					"groups":     strings.Join(groups, "."),
				},
			})
			result = append(result, r)
		}
	}
	// policy attachments are not resources
	// what happens here, is that policy is detached from groups, users and roles
	return []Resources{resultAtt, result}
}

func filterIamRole(res Resources, raw interface{}, f Filter, c *AWSClient) []Resources {
	ids := []*string{}
	rpolIds := []*string{}
	rpolAttributes := []*map[string]string{}
	pIds := []*string{}

	for _, role := range raw.(*iam.ListRolesOutput).Roles {
		if f.Matches(res[0].Type, *role.RoleName) {
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
	//{Type: "aws_iam_role_policy_attachment", Ids: rpolIds, Attrs: rpolAttributes},
	//{Type: "aws_iam_role_policy", Ids: pIds},
	//{Type: res.Type, Ids: ids},
	}
}

func filterInstanceProfiles(res Resources, raw interface{}, f Filter, c *AWSClient) []Resources {
	ids := []*string{}
	attributes := []*map[string]string{}

	for _, r := range raw.(*iam.ListInstanceProfilesOutput).InstanceProfiles {
		if f.Matches(res[0].Type, *r.InstanceProfileName) {
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
	return []Resources{
	//{Type: res.Type, Ids: ids}
	}
}

func filterKmsKeys(res Resources, raw interface{}, f Filter, c *AWSClient) []Resources {
	ids := []*string{}
	attributes := []*map[string]string{}

	for _, r := range raw.(*kms.ListKeysOutput).Keys {
		if f.Matches(res[0].Type, *r.KeyArn) {
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
	return []Resources{
	//{Type: res.Type, Ids: ids}
	}
}
