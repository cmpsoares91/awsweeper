package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/iflix/awsweeper/command"
	res "github.com/iflix/awsweeper/resource"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/spf13/afero"
)

func TestAccIamPolicy_deleteByIds(t *testing.T) {
	var p1, p2 iam.Policy

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccIamPolicyConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIamPolicyExists("aws_iam_policy.foo", &p1),
					testAccCheckIamPolicyExists("aws_iam_policy.bar", &p2),
					testMainIamPolicyIds(argsDryRun, &p1),
					testIamPolicyExists(&p1),
					testIamPolicyExists(&p2),
					testMainIamPolicyIds(argsForceDelete, &p1),
					testIamPolicyDeleted(&p1),
					testIamPolicyExists(&p2),
				),
			},
		},
	})
}

func TestAccIamPolicyAttached_deleteByIds(t *testing.T) {
	var p1, p2 iam.Policy

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccIamPolicyAttachedConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIamPolicyExists("aws_iam_policy.foo", &p1),
					testAccCheckIamPolicyExists("aws_iam_policy.bar", &p2),
					testMainIamPolicyIds(argsDryRun, &p1),
					testIamPolicyExists(&p1),
					testIamPolicyExists(&p2),
					testMainIamPolicyIds(argsForceDelete, &p1),
					testIamPolicyDeleted(&p1),
					testIamPolicyExists(&p2),
				),
			},
		},
	})
}

func testMainIamPolicyIds(args []string, p *iam.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res.AppFs = afero.NewMemMapFs()
		afero.WriteFile(res.AppFs, "config.yml", []byte(testAWSweeperIdsConfig(res.IamPolicy, p.Arn)), 0644)
		os.Args = args
		command.WrappedMain()
		return nil
	}
}

func testAccCheckIamPolicyExists(name string, p *iam.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		conn := client.IAMAPI
		desc := &iam.GetPolicyInput{
			PolicyArn: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetPolicy(desc)
		if err != nil {
			iamErr, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if iamErr.Code() == NoSuchEntity {
				return fmt.Errorf("IAM policy has been deleted")
			}
			return err
		}

		*p = *resp.Policy

		return nil
	}
}

func testIamPolicyExists(p *iam.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.IAMAPI
		desc := &iam.GetPolicyInput{
			PolicyArn: p.Arn,
		}
		_, err := conn.GetPolicy(desc)
		if err != nil {
			iamErr, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if iamErr.Code() == NoSuchEntity {
				return fmt.Errorf("IAM policy has been deleted")
			}
			return err
		}

		return nil
	}
}

func testIamPolicyDeleted(p *iam.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.IAMAPI

		desc := &iam.GetPolicyInput{
			PolicyArn: p.Arn,
		}
		_, err := conn.GetPolicy(desc)
		if err != nil {
			iamErr, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if iamErr.Code() == NoSuchEntity {
				return nil
			}
			return err
		}
		return fmt.Errorf("IAM policy hasn't been deleted")
	}
}

const testAccIamPolicyConfig = `
resource "aws_iam_policy" "foo" {
  name        = "foo"
  path        = "/awsweeper-testacc/"
  description = "My foo test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "bar" {
  name        = "bar"
  path        = "/awsweeper-testacc/"
  description = "My bar test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}
`

const testAccIamPolicyAttachedConfig = `
resource "aws_iam_policy" "foo" {
  name        = "foo"
  path        = "/awsweeper-testacc/"
  description = "My foo test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "bar" {
  name        = "bar"
  path        = "/awsweeper-testacc/"
  description = "My bar test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_user" "user" {
  name = "test-user"
  path = "/awsweeper-testacc/"
}

resource "aws_iam_role" "role" {
  name = "test-role"
  path = "/awsweeper-testacc/"
  assume_role_policy = "${data.aws_iam_policy_document.test-assume-role-policy.json}"
}

data "aws_iam_policy_document" "test-assume-role-policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_group" "group" {
  name = "test-group"
  path = "/awsweeper-testacc/"
}

resource "aws_iam_policy_attachment" "test-attach" {
  name       = "awsweeper-testacc-policy-attachment"
  users      = ["${aws_iam_user.user.name}"]
  roles      = ["${aws_iam_role.role.name}"]
  groups     = ["${aws_iam_group.group.name}"]
  policy_arn = "${aws_iam_policy.foo.arn}"
}
`
