package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/cloudetc/awsweeper/command"
	"github.com/spf13/afero"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/route53"
	res "github.com/cloudetc/awsweeper/resource"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccRoute53Zone_deleteByTags(t *testing.T) {
	// TODO tags are a special case for this resource and are not supported yet
	t.Skip("Costs money even in free tier")
	var zone1, zone2 route53.HostedZone

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: sharedTfAwsProvider,
		Steps: []resource.TestStep{
			{
				Config:             testAccRoute53ZoneConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists("aws_route53_zone.foo", &zone1),
					testAccCheckRoute53ZoneExists("aws_route53_zone.bar", &zone2),
					testMainTags(argsDryRun, testAWSweeperTagsConfig(res.Route53Zone)),
					testRoute53ZoneExists(&zone1),
					testRoute53ZoneExists(&zone2),
					testMainTags(argsForceDelete, testAWSweeperTagsConfig(res.Route53Zone)),
					testRoute53ZoneDeleted(&zone1),
					testRoute53ZoneExists(&zone2),
				),
			},
		},
	})
}

func TestAccRoute53Zone_deleteByIds(t *testing.T) {
	t.Skip("Costs money even in free tier")
	var zone1, zone2 route53.HostedZone

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: sharedTfAwsProvider,
		Steps: []resource.TestStep{
			{
				Config:             testAccRoute53ZoneConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists("aws_route53_zone.foo", &zone1),
					testAccCheckRoute53ZoneExists("aws_route53_zone.bar", &zone2),
					testMainRoute53ZoneIds(argsDryRun, &zone1),
					testRoute53ZoneExists(&zone1),
					testRoute53ZoneExists(&zone2),
					testMainRoute53ZoneIds(argsForceDelete, &zone1),
					testRoute53ZoneDeleted(&zone1),
					testRoute53ZoneExists(&zone2),
				),
			},
		},
	})
}

func testMainRoute53ZoneIds(args []string, z *route53.HostedZone) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res.AppFs = afero.NewMemMapFs()
		afero.WriteFile(res.AppFs, "config.yml", []byte(testAWSweeperIdsConfig(res.Route53Zone, z.Id)), 0644)
		os.Args = args
		command.WrappedMain()
		return nil
	}
}

func testAccCheckRoute53ZoneExists(n string, z *route53.HostedZone) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		conn := sharedAwsClient.Route53API
		desc := &route53.GetHostedZoneInput{
			Id: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetHostedZone(desc)
		if err != nil {
			route53err, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if route53err.Code() == NoSuchHostedZone {
				return nil
			}
			return err
		}

		*z = *resp.HostedZone

		return nil
	}
}

func testRoute53ZoneExists(z *route53.HostedZone) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := sharedAwsClient.Route53API
		desc := &route53.GetHostedZoneInput{
			Id: z.Id,
		}
		_, err := conn.GetHostedZone(desc)
		if err != nil {
			route53err, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if route53err.Code() == NoSuchHostedZone {
				return fmt.Errorf("route53 zone has been deleted")
			}
			return err
		}

		return nil
	}
}

func testRoute53ZoneDeleted(z *route53.HostedZone) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := sharedAwsClient.Route53API
		desc := &route53.GetHostedZoneInput{
			Id: z.Id,
		}
		_, err := conn.GetHostedZone(desc)
		if err != nil {
			route53err, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if route53err.Code() == NoSuchHostedZone {
				return nil
			}
			return err
		}
		return fmt.Errorf("route53 Zone hasn't been deleted")
	}
}

const testAccRoute53ZoneConfig = `
resource "aws_route53_zone" "foo" {
	name = "foo.com"

tags {
		foo = "bar"
		Name = "awsweeper-testacc"
	}
}

resource "aws_route53_zone" "bar" {
	name = "bar.com"

	tags {
		bar = "baz"
		Name = "awsweeper-testacc"
	}
}

resource "aws_route53_record" "foo" {
  zone_id = "${aws_route53_zone.foo.zone_id}"
  name    = "bar.com"
  type    = "NS"
  ttl     = "30"

  records = [
    "${aws_route53_zone.bar.name_servers.0}",
    "${aws_route53_zone.bar.name_servers.1}",
    "${aws_route53_zone.bar.name_servers.2}",
    "${aws_route53_zone.bar.name_servers.3}",
  ]
}
`
