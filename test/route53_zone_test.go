package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/cloudetc/awsweeper/command"
	res "github.com/cloudetc/awsweeper/resource"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/spf13/afero"
)

func TestAccRoute53Zone_deleteByTags(t *testing.T) {
	// TODO tags are a special case for this resource and are not supported yet
	t.SkipNow()
	var zone1, zone2 route53.HostedZone

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccRoute53ZoneConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists("aws_route53_zone.foo", &zone1),
					testAccCheckRoute53ZoneExists("aws_route53_zone.bar", &zone2),
					testMainTags(argsDryRun, testAccRoute53ZoneAWSweeperTagsConfig),
					testRoute53ZoneExists(&zone1),
					testRoute53ZoneExists(&zone2),
					testMainTags(argsForceDelete, testAccRoute53ZoneAWSweeperTagsConfig),
					testRoute53ZoneDeleted(&zone1),
					testRoute53ZoneExists(&zone2),
				),
			},
		},
	})
}

func TestAccRoute53Zone_deleteByIds(t *testing.T) {
	var zone1, zone2 route53.HostedZone

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
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

func testAccCheckRoute53ZoneExists(n string, z *route53.HostedZone) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := client.R53conn
		desc := &route53.GetHostedZoneInput{
			Id: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetHostedZone(desc)
		if err != nil {
			route53err, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if route53err.Code() == "NoSuchHostedZone" {
				return nil
			}
			return err
		}

		*z = *resp.HostedZone

		return nil
	}
}

func testMainRoute53ZoneIds(args []string, z *route53.HostedZone) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res.AppFs = afero.NewMemMapFs()
		afero.WriteFile(res.AppFs, "config.yml", []byte(testAccRoute53ZoneAWSweeperIdsConfig(z)), 0644)
		os.Args = args

		command.WrappedMain()
		return nil
	}
}

func testRoute53ZoneExists(z *route53.HostedZone) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.R53conn
		desc := &route53.GetHostedZoneInput{
			Id: z.Id,
		}
		_, err := conn.GetHostedZone(desc)
		if err != nil {
			route53err, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if route53err.Code() == "NoSuchHostedZone" {
				return fmt.Errorf("Route53 Zone has been deleted")
			}
			return err
		}

		return nil
	}
}

func testRoute53ZoneDeleted(z *route53.HostedZone) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.R53conn
		desc := &route53.GetHostedZoneInput{
			Id: z.Id,
		}
		_, err := conn.GetHostedZone(desc)
		if err != nil {
			route53err, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if route53err.Code() == "NoSuchHostedZone" {
				return nil
			}
			return err
		}
		return fmt.Errorf("Route53 Zone hasn't been deleted")
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

const testAccRoute53ZoneAWSweeperTagsConfig = `
aws_route53_zone:
  tags:
    foo: bar
`

func testAccRoute53ZoneAWSweeperIdsConfig(z *route53.HostedZone) string {
	id := z.Id
	return fmt.Sprintf(`
aws_route53_zone:
  ids:
    - %s
`, *id)
}
