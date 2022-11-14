package aws

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

func createOrUpdateIAM(ctx context.Context, env string, slug string) auto.Stack {
	project := slug
	stackName := fmt.Sprintf("%s-%s-iamRole", project, env)

	return createOrSelectStack(ctx, project, stackName, iamRoleFunc)
}

func iamRoleFunc(ctx pulumi.Context) error {
	stack := ctx.Stack()
	project := ctx.Project()

	iamrole, err := iam.NewRole(ctx, fmt.Sprintf(""), &iam.RoleArgs{
		Description: fmt.Sprintf("IAM Role for %s", stack),
		RoleName:    fmt.Sprintf("fh-%s", stack),

		Tags: []iam.RoleTagArgs{
			{
				Key:   pulumi.String("stack"),
				Value: pulumi.String(fmt.Sprintf("%s/%s", project, stack)),
			},
		},
	})
	if err != nil {
		return err
	}

	// Export Iam role Arn
	ctx.Export("iamRole", iamrole.Arn)
	return nil
}
