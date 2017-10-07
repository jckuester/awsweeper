package main

import (
	"strings"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform/terraform"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/route53"
	"fmt"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kms"
	"regexp"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"github.com/mitchellh/cli"
	"os"
	"sync"
)

type yamlCfg struct {
	Ids  []*string `yaml:",omitempty"`
	Tags map[string]string `yaml:",omitempty"`
}

type WipeCommand struct {
	Ui       		cli.Ui
	isTestRun		bool
	client			*AWSClient
	provider        *terraform.ResourceProvider
	resourceTypes   []string
	filter          []*ec2.Filter
	deleteCfg	    map[string]yamlCfg
	deleteOut		map[string]yamlCfg
	outFileName		string
}

type Resources struct {
	Type  string
	Ids   []*string
	Attrs []*map[string]string
	Tags  []*map[string]string
}

type Resource struct {
	id *string
	attrs *map[string]string
	tags *map[string]string
}

type AWSClient struct {
	ec2conn         *ec2.EC2
	autoscalingconn *autoscaling.AutoScaling
	elbconn         *elb.ELB
	r53conn         *route53.Route53
	cfconn          *cloudformation.CloudFormation
	efsconn         *efs.EFS
	iamconn         *iam.IAM
	kmsconn         *kms.KMS
}

func (c *WipeCommand) Run(args []string) int {
	c.deleteCfg = map[string]yamlCfg{}
	c.deleteOut = map[string]yamlCfg{}

	c.resourceTypes = []string{
		"aws_autoscaling_group",
		"aws_launch_configuration",
		"aws_instance",
		"aws_elb",
		"aws_vpc_endpoint",
		"aws_nat_gateway",
		"aws_cloudformation_stack",
		"aws_route53_zone",
		"aws_eip",
		"aws_internet_gateway",
		"aws_efs_file_system",
		"aws_network_interface",
		"aws_subnet",
		"aws_route_table",
		"aws_network_acl",
		"aws_security_group",
		"aws_vpc",
		"aws_iam_user",
		"aws_iam_role",
		"aws_iam_policy",
		"aws_iam_instance_profile",
		"aws_kms_alias",
		"aws_kms_key",
	}

	deleteFunctions := map[string]func(string){
		"aws_autoscaling_group":    c.deleteASGs,
		"aws_launch_configuration": c.deleteLCs,
		"aws_instance":             c.deleteInstances,
		"aws_internet_gateway":     c.deleteInternetGateways,
		"aws_eip":                  c.deleteEips,
		"aws_elb":                  c.deleteELBs,
		"aws_vpc_endpoint":         c.deleteVpcEndpoints,
		"aws_nat_gateway":          c.deleteNatGateways,
		"aws_network_interface":    c.deleteNetworkInterfaces,
		"aws_route_table":          c.deleteRouteTables,
		"aws_security_group":       c.deleteSecurityGroups,
		"aws_network_acl":          c.deleteNetworkAcls,
		"aws_subnet":               c.deleteSubnets,
		"aws_cloudformation_stack": c.deleteCloudformationStacks,
		"aws_route53_zone":         c.deleteRoute53Zone,
		"aws_vpc":                  c.deleteVpcs,
		"aws_efs_file_system":      c.deleteEfsFileSystem,
		"aws_iam_user":             c.deleteIamUser,
		"aws_iam_role":             c.deleteIamRole,
		"aws_iam_policy":           c.deleteIamPolicy,
		"aws_iam_instance_profile": c.deleteInstanceProfiles,
		"aws_kms_alias":            c.deleteKmsAliases,
		"aws_kms_key":              c.deleteKmsKeys,
	}

	if len(args) == 1 {
		data, err := ioutil.ReadFile(args[0])
		check(err)
		err = yaml.Unmarshal([]byte(data), &c.deleteCfg)
		check(err)
	} else {
		fmt.Println(Help())
		return 1
	}

	if c.isTestRun {
		c.Ui.Output("INFO: This is a test run, nothing will be deleted!")
	}

	for _, k := range c.resourceTypes {
		deleteFunctions[k](k)
	}

	if c.outFileName != "" {
		outYaml, err := yaml.Marshal(&c.deleteOut)
		check(err)

		fileYaml := []byte(string(outYaml))
		err = ioutil.WriteFile(c.outFileName, fileYaml, 0644)
		check(err)
	}

	return 0
}

func (c *WipeCommand) Help() string {
	return Help()
}

func (c *WipeCommand) Synopsis() string {
	return "Delete all or one specific resource type"
}

func (c *WipeCommand) deleteASGs(resourceType string) {
	res, err := c.client.autoscalingconn.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	check(err)
	
	ids := []*string{}
	tags := []*map[string]string{}

	for _, r := range res.AutoScalingGroups {
		if c.checkDelete(resourceType, r.AutoScalingGroupName) {
			ids = append(ids, r.AutoScalingGroupName)

			m := map[string]string{}
			for _, t := range r.Tags {
				m[*t.Key] = *t.Value
			}
			tags = append(tags, &m)
		}
	}

	c.delete(Resources{Type: resourceType, Ids: ids, Tags: tags})
}

func (c *WipeCommand) deleteLCs(resourceType string) {
	res, err := c.client.autoscalingconn.DescribeLaunchConfigurations(&autoscaling.DescribeLaunchConfigurationsInput{})
	check(err)

	ids := []*string{}

	for _, r := range res.LaunchConfigurations {
		if c.checkDelete(resourceType, r.LaunchConfigurationName) {
			ids = append(ids, r.LaunchConfigurationName)
		}
	}
	c.delete(Resources{Type: resourceType, Ids: ids})
}

func (c *WipeCommand) deleteInstances(resourceType string) {
	res, err := c.client.ec2conn.DescribeInstances(&ec2.DescribeInstancesInput{})
	check(err)

	ids := []*string{}
	tags := []*map[string]string{}

	for _, r := range res.Reservations {
		for _, in := range r.Instances {
			if *in.State.Name != "terminated" {
				m := &map[string]string{}
				for _, t := range in.Tags {
					(*m)[*t.Key] = *t.Value
				}

				if c.checkDelete(resourceType, in.InstanceId, m) {
					ids = append(ids, in.InstanceId)
					tags = append(tags, m)
				}
			}
		}
	}
	c.delete(Resources{Type: resourceType, Ids: ids, Tags: tags})
}

func (c *WipeCommand) deleteInternetGateways(resourceType string) {
	res, err := c.client.ec2conn.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{})
	check(err)

	ids := []*string{}
	attrs := []*map[string]string{}
	tags := []*map[string]string{}

	for _, r := range res.InternetGateways {
		m := &map[string]string{}
		for _, t := range r.Tags {
			(*m)[*t.Key] = *t.Value
		}

		if c.checkDelete(resourceType, r.InternetGatewayId, m) {
			ids = append(ids, r.InternetGatewayId)
			attrs = append(attrs, &map[string]string{
				"vpc_id": *r.Attachments[0].VpcId,
			})
			tags = append(tags, m)
		}
	}
	c.delete(Resources{Type: resourceType, Ids: ids, Attrs: attrs, Tags: tags})
}

func (c *WipeCommand) deleteNatGateways(resourceType string) {
	res, err := c.client.ec2conn.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{})
	check(err)

	ids := []*string{}
	for _, r := range res.NatGateways {
		if c.checkDelete(resourceType, r.NatGatewayId) {
			if *r.State == "available" {
				ids = append(ids, r.NatGatewayId)
			}
		}
	}
	c.delete(Resources{Type: resourceType, Ids: ids})
}

func (c *WipeCommand) deleteRouteTables(resourceType string) {
	res, err := c.client.ec2conn.DescribeRouteTables(&ec2.DescribeRouteTablesInput{})
	check(err)

	ids := []*string{}
	tags := []*map[string]string{}

	for _, r := range res.RouteTables {
		m := &map[string]string{}
		for _, t := range r.Tags {
			(*m)[*t.Key] = *t.Value
		}

		if c.checkDelete(resourceType, r.RouteTableId, m) {
			main := false
			for _, a := range r.Associations {
				if *a.Main {
					main = true
				}
			}
			if ! main {
				ids = append(ids, r.RouteTableId)
				tags = append(tags, m)
			}
		}
	}
	// aws_route_table_association handled implicitly
	c.delete(Resources{Type: resourceType, Ids: ids, Tags: tags})
}

func (c *WipeCommand) deleteSecurityGroups(resourceType string) {
	res, err := c.client.ec2conn.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})
	check(err)

	ids := []*string{}
	tags := []*map[string]string{}

	for _, r := range res.SecurityGroups {
		m := &map[string]string{}
		for _, t := range r.Tags {
			(*m)[*t.Key] = *t.Value
		}

		if c.checkDelete(resourceType, r.GroupId, m) {
			if *r.GroupName != "default" {
				ids = append(ids, r.GroupId)
				tags = append(tags, m)
			}
		}
	}
	c.delete(Resources{Type: resourceType, Ids: ids, Tags: tags})
}

func (c *WipeCommand) deleteNetworkAcls(resourceType string) {
	res, err := c.client.ec2conn.DescribeNetworkAcls(&ec2.DescribeNetworkAclsInput{})
	check(err)

	ids := []*string{}
	tags := []*map[string]string{}

	for _, r := range res.NetworkAcls {
		m := &map[string]string{}
		for _, t := range r.Tags {
			(*m)[*t.Key] = *t.Value
		}

		if ! *r.IsDefault {
			if c.checkDelete(resourceType, r.NetworkAclId, m) {
				ids = append(ids, r.NetworkAclId)
				// TODO handle associations
				tags = append(tags, m)
			}
		}
	}
	c.delete(Resources{Type: resourceType, Ids: ids, Tags: tags})
}

func (c *WipeCommand) deleteNetworkInterfaces(resourceType string) {
	res, err := c.client.ec2conn.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{})
	check(err)

	ids := []*string{}
	tags := []*map[string]string{}

	for _, r := range res.NetworkInterfaces {
		m := &map[string]string{}
		for _, t := range r.TagSet {
			(*m)[*t.Key] = *t.Value
		}
		if c.checkDelete(resourceType, r.NetworkInterfaceId, m) {
			ids = append(ids, r.NetworkInterfaceId)
			tags = append(tags, m)
		}
	}
	c.delete(Resources{Type: resourceType, Ids: ids, Tags: tags})
}

func (c *WipeCommand) deleteELBs(resourceType string) {
	res, err := c.client.elbconn.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})
	check(err)

	ids := []*string{}
	for _, r := range res.LoadBalancerDescriptions {
		if c.checkDelete(resourceType, r.LoadBalancerName) {
			ids = append(ids, r.LoadBalancerName)
		}
	}
	c.delete(Resources{Type: resourceType, Ids: ids})
}

func (c *WipeCommand) deleteVpcEndpoints(resourceType string) {
	res, err := c.client.ec2conn.DescribeVpcEndpoints(&ec2.DescribeVpcEndpointsInput{})
	check(err)

	ids := []*string{}
	for _, r := range res.VpcEndpoints {
		if c.checkDelete(resourceType, r.VpcEndpointId) {
			ids = append(ids, r.VpcEndpointId)
		}
	}
	c.delete(Resources{Type: resourceType, Ids: ids})
}

func (c *WipeCommand) deleteEips(resourceType string) {
	res, err := c.client.ec2conn.DescribeAddresses(&ec2.DescribeAddressesInput{})
	check(err)

	ids := []*string{}
	for _, r := range res.Addresses {
		if c.checkDelete(resourceType, r.AllocationId) {
			ids = append(ids, r.AllocationId)
		}
	}
	c.delete(Resources{Type: resourceType, Ids: ids})
}

func (c *WipeCommand) deleteSubnets(resourceType string) {
	res, err := c.client.ec2conn.DescribeSubnets(&ec2.DescribeSubnetsInput{})
	check(err)

	ids := []*string{}
	tags := []*map[string]string{}

	for _, r := range res.Subnets {
		m := &map[string]string{}
		for _, t := range r.Tags {
			(*m)[*t.Key] = *t.Value
		}
		if c.checkDelete(resourceType, r.SubnetId, m) {
			ids = append(ids, r.SubnetId)
			tags = append(tags, m)
		}
	}
	c.delete(Resources{Type: resourceType, Ids: ids, Tags: tags})
}

func (c *WipeCommand) deleteVpcs(resourceType string) {
	res, err := c.client.ec2conn.DescribeVpcs(&ec2.DescribeVpcsInput{})
	check(err)

	ids := []*string{}
	tags := []*map[string]string{}

	for _, r := range res.Vpcs {
		m := &map[string]string{}
		for _, t := range r.Tags {
			(*m)[*t.Key] = *t.Value
		}

		if c.checkDelete(resourceType, r.VpcId, m) {
			ids = append(ids, r.VpcId)
			tags = append(tags, m)
		}
	}
	c.delete(Resources{Type: resourceType, Ids: ids, Tags: tags})
}

func (c *WipeCommand) deleteRoute53Record(resourceType string) {
	res, err := c.client.r53conn.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{})
	check(err)

	ids := []*string{}
	for _, r := range res.ResourceRecordSets {
		for _, rr := range r.ResourceRecords {
			if c.checkDelete(resourceType, rr.Value) {
				ids = append(ids, rr.Value)
			}
		}
	}
	c.delete(Resources{Type: resourceType, Ids: ids})
}

func (c *WipeCommand) deleteRoute53Zone(resourceType string) {
	res, err := c.client.r53conn.ListHostedZones(&route53.ListHostedZonesInput{})
	check(err)

	hzIds := []*string{}
	rsIds := []*string{}
	rsAttrs := []*map[string]string{}
	hzAttrs := []*map[string]string{}

	for _, hz := range res.HostedZones {
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
	c.delete(Resources{Type: "aws_route53_record", Ids: rsIds, Attrs: rsAttrs})
	c.delete(Resources{Type: resourceType, Ids: hzIds, Attrs: hzAttrs})
}

func (c *WipeCommand) deleteCloudformationStacks(resourceType string) {
	res, err := c.client.cfconn.DescribeStacks(&cloudformation.DescribeStacksInput{})
	check(err)

	ids := []*string{}
	tags := []*map[string]string{}

	for _, r := range res.Stacks {
		m := &map[string]string{}
		for _, t := range r.Tags {
			(*m)[*t.Key] = *t.Value
		}

		if c.checkDelete(resourceType, r.StackId, m) {
			// TODO filter name?
			ids = append(ids, r.StackId)
			tags = append(tags, m)
		}
	}
	c.delete(Resources{Type: resourceType, Ids: ids, Tags: tags})
}

func (c *WipeCommand) deleteEfsFileSystem(resourceType string) {
	res, err := c.client.efsconn.DescribeFileSystems(&efs.DescribeFileSystemsInput{})
	check(err)

	fsIds := []*string{}
	mtIds := []*string{}

	for _, r := range res.FileSystems {
		if c.checkDelete(resourceType, r.Name) {
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
	c.delete(Resources{Type: "aws_efs_mount_target", Ids: mtIds})
	c.delete(Resources{Type: resourceType, Ids: fsIds})
}

func (c *WipeCommand) deleteIamUser(resourceType string) {
	users, err := c.client.iamconn.ListUsers(&iam.ListUsersInput{})
	check(err)

	ids := []*string{}
	pIds := []*string{}
	attrs := []*map[string]string{}
	pAttrs := []*map[string]string{}

	for _, u := range users.Users {
		if c.checkDelete(resourceType, u.UserName) {
			ups, err := c.client.iamconn.ListUserPolicies(&iam.ListUserPoliciesInput{
				UserName: u.UserName,
			})
			if err == nil {
				for _, up := range ups.PolicyNames {
					fmt.Println(*up)
				}
			}

			upols, err := c.client.iamconn.ListAttachedUserPolicies(&iam.ListAttachedUserPoliciesInput{
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
	c.delete(Resources{Type: "aws_iam_user_policy_attachment", Ids: pIds, Attrs: pAttrs})
	c.delete(Resources{Type: resourceType, Ids: ids, Attrs: attrs})
}

func (c *WipeCommand) deleteIamPolicy(resourceType string) {
	ps, err := c.client.iamconn.ListPolicies(&iam.ListPoliciesInput{})
	check(err)

	//ps, err := c.client.iamconn.ListGroups(&iam.ListPoliciesInput{})

	ids := []*string{}
	eIds := []*string{}
	attributes := []*map[string]string{}

	for _, pol := range ps.Policies {
		if c.checkDelete(resourceType, pol.Arn) {
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
	c.delete(Resources{Type: "aws_iam_policy_attachment", Ids: eIds, Attrs: attributes})
	c.delete(Resources{Type: resourceType, Ids: ids})
}

func (c *WipeCommand) deleteIamRole(resourceType string) {
	roles, err := c.client.iamconn.ListRoles(&iam.ListRolesInput{})
	check(err)

	ids := []*string{}
	rpolIds := []*string{}
	rpolAttributes := []*map[string]string{}
	pIds := []*string{}

	for _, role := range roles.Roles {
		if c.checkDelete(resourceType, role.RoleName) {
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

			ips, err := c.client.iamconn.ListInstanceProfilesForRole(&iam.ListInstanceProfilesForRoleInput{
				RoleName: role.RoleName,
			})
			check(err)

			for _, ip := range ips.InstanceProfiles {
				fmt.Println(ip.InstanceProfileName)
			}

			ids = append(ids, role.RoleName)
		}
	}
	c.delete(Resources{Type: "aws_iam_role_policy_attachment", Ids: rpolIds, Attrs: rpolAttributes})
	c.delete(Resources{Type: "aws_iam_role_policy", Ids: pIds})
	c.delete(Resources{Type: resourceType, Ids: ids})
}

func (c *WipeCommand) deleteInstanceProfiles(resourceType string) {
	res, err := c.client.iamconn.ListInstanceProfiles(&iam.ListInstanceProfilesInput{})
	check(err)

	ids := []*string{}
	attributes := []*map[string]string{}

	for _, r := range res.InstanceProfiles {
		if c.checkDelete(resourceType, r.InstanceProfileName) {
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
	c.delete(Resources{Type: resourceType, Ids: ids})
}

func (c *WipeCommand) deleteKmsAliases(resourceType string) {
	res, err := c.client.kmsconn.ListAliases(&kms.ListAliasesInput{})
	check(err)

	ids := []*string{}

	for _, r := range res.Aliases {
		if c.checkDelete(resourceType, r.AliasArn) {
			ids = append(ids, r.AliasArn)
		}
	}
	c.delete(Resources{Type: resourceType, Ids: ids})
}

func (c *WipeCommand) deleteKmsKeys(resourceType string) {
	res, err := c.client.kmsconn.ListKeys(&kms.ListKeysInput{})
	check(err)

	ids := []*string{}
	attributes := []*map[string]string{}

	for _, r := range res.Keys {
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
	c.delete(Resources{Type: resourceType, Ids: ids})
}

func (c *WipeCommand) checkDelete(rType string, id *string, tags ...*map[string]string) bool {
	if rVal, ok := c.deleteCfg[rType]; ok {
		if len(rVal.Ids) == 0 && len(rVal.Tags) == 0 {
			return true
		}
		for _, regex := range rVal.Ids {
			res, _ := regexp.MatchString(*regex, *id)
			if res {
				return true
			}
		}
		for k, v := range rVal.Tags {
			if len(tags) > 0 {
				t := tags[0]
				if tVal, ok := (*t)[k]; ok {
					res, _ := regexp.MatchString(v, tVal)
					if res {
						return true
					}
				}
			}
		}
	}
	return false
}

func (c *WipeCommand) delete(res Resources) {
	numWorkerThreads := 10

	if len(res.Ids) == 0 {
		return
	}

	c.deleteOut[res.Type] = yamlCfg{Ids: res.Ids}

	fmt.Printf("\n---\nType: %s\nFound: %d\n\n", res.Type, len(res.Ids))

	ii := &terraform.InstanceInfo{
		Type: res.Type,
	}

	d := &terraform.InstanceDiff{
		Destroy: true,
	}

	a := []*map[string]string{}
	if len(res.Attrs) > 0 {
		a = res.Attrs
	} else {
		for i := 0; i < len(res.Ids); i++ {
			a = append(a, &map[string]string{})
		}
	}

	ts := make([]*map[string]string, len(res.Ids))
	if len(res.Tags) > 0 {
		ts = res.Tags
	}
	chResources := make(chan *Resource, numWorkerThreads)

	var wg sync.WaitGroup
	wg.Add(len(res.Ids))

	for j := 1; j <= numWorkerThreads; j++ {
		go func() {
			for {
				res, more := <- chResources
				if more {
					printStat := fmt.Sprintf("\tId:\t%s", *res.id)
					if res.tags != nil {
						if len(*res.tags) > 0 {
							printStat += "\n\tTags:\t"
							for k, v := range *res.tags {
								printStat += fmt.Sprintf("[%s: %v] ", k, v)
							}
							printStat += "\n"
						}
					}
					fmt.Println(printStat)

					a := res.attrs
					(*a)["force_destroy"] = "true"

					s := &terraform.InstanceState{
						ID: *res.id,
						Attributes: *a,
					}

					st, err := (*c.provider).Refresh(ii, s)
					if err != nil{
						fmt.Println("err: ", err)
						st = s
						st.Attributes["force_destroy"] = "true"
					}

					if !c.isTestRun {
						_, err := (*c.provider).Apply(ii, st, d)

						if err != nil {
							fmt.Printf("\t%s\n", err)
						}
					}
					wg.Done()
				} else {
					return
				}
			}
		}()
	}

	for i, id := range res.Ids {
		if id != nil {
			chResources <- &Resource{
				id: id,
				attrs: a[i],
				tags: ts[i],
			}
		}
	}
	close(chResources)

	wg.Wait()
	fmt.Println("---\n")
}

func check(e error) {
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
		//panic(e)
	}
}
