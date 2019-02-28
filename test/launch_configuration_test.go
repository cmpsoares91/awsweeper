package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/iflix/awsweeper/command"
	res "github.com/iflix/awsweeper/resource"
	"github.com/spf13/afero"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccLaunchConfiguration_deleteByIds(t *testing.T) {
	var lc1, lc2 autoscaling.LaunchConfiguration

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccLaunchConfigurationConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists("aws_launch_configuration.foo", &lc1),
					testAccCheckLaunchConfigurationExists("aws_launch_configuration.bar", &lc2),
					testMainLaunchConfigurationIds(argsDryRun, &lc1),
					testLaunchConfigurationExists(&lc1),
					testLaunchConfigurationExists(&lc2),
					testMainLaunchConfigurationIds(argsForceDelete, &lc1),
					testLaunchConfigurationDeleted(&lc1),
					testLaunchConfigurationExists(&lc2),
				),
			},
		},
	})
}

func testMainLaunchConfigurationIds(args []string, lc *autoscaling.LaunchConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res.AppFs = afero.NewMemMapFs()
		afero.WriteFile(res.AppFs, "config.yml",
			[]byte(testAWSweeperIdsConfig(res.LaunchConfiguration, lc.LaunchConfigurationName)), 0644)
		os.Args = args
		command.WrappedMain()
		return nil
	}
}

func testAccCheckLaunchConfigurationExists(n string, lc *autoscaling.LaunchConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Launch Configuration name is set")
		}

		conn := client.AutoScalingAPI
		DescribeLaunchConfigurationOpts := &autoscaling.DescribeLaunchConfigurationsInput{
			LaunchConfigurationNames: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeLaunchConfigurations(DescribeLaunchConfigurationOpts)
		if err != nil {
			return err
		}
		if len(resp.LaunchConfigurations) == 0 {
			return fmt.Errorf("launch Configuration not found")
		}

		*lc = *resp.LaunchConfigurations[0]

		return nil
	}
}

func testLaunchConfigurationDeleted(lc *autoscaling.LaunchConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.AutoScalingAPI
		DescribeLaunchConfigurationOpts := &autoscaling.DescribeLaunchConfigurationsInput{
			LaunchConfigurationNames: []*string{lc.LaunchConfigurationName},
		}
		resp, err := conn.DescribeLaunchConfigurations(DescribeLaunchConfigurationOpts)
		if err != nil {
			ec2err, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if ec2err.Code() == "InvalidLaunchConfiguration.NotFound" {
				return nil
			}
			return err
		}

		if len(resp.LaunchConfigurations) != 0 {
			return fmt.Errorf("launch Configuration hasn't been deleted")
		}

		return nil
	}
}

func testLaunchConfigurationExists(lc *autoscaling.LaunchConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.AutoScalingAPI
		DescribeLaunchConfigurationOpts := &autoscaling.DescribeLaunchConfigurationsInput{
			LaunchConfigurationNames: []*string{lc.LaunchConfigurationName},
		}
		resp, err := conn.DescribeLaunchConfigurations(DescribeLaunchConfigurationOpts)
		if err != nil {
			return err
		}
		if len(resp.LaunchConfigurations) == 0 {
			return fmt.Errorf("launch Configuration has been deleted")
		}

		return nil
	}
}

const testAccLaunchConfigurationConfig = `
resource "aws_launch_configuration" "foo" {
	name_prefix = "awsweeper-testacc-foo-"
	image_id = "${data.aws_ami.foo.id}"
	instance_type = "t2.micro"

	lifecycle {
		create_before_destroy = true
	}
}

resource "aws_launch_configuration" "bar" {
	name_prefix = "awsweeper-testacc-bar-"
	image_id = "${data.aws_ami.foo.id}"
	instance_type = "t2.micro"

	lifecycle {
		create_before_destroy = true
	}
}

data "aws_ami" "foo" {
	most_recent = true
	owners = ["099720109477"]

	filter {
		name = "name"
		values = ["*ubuntu-trusty-14.04-amd64-server-*"]
	}

	filter {
		name = "state"
		values = ["available"]
	}

	filter {
		name = "virtualization-type"
		values = ["hvm"]
	}

	filter {
		name = "is-public"
		values = ["true"]
	}
}
`
