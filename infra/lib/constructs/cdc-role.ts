import { Construct } from "constructs";
import { RoleProps } from "../types/role-props";
import * as iam from "aws-cdk-lib/aws-iam";
import { Fn } from "aws-cdk-lib";
import { getAccountId } from "../utils/account-utils";

export class CdcRole extends Construct {
  private readonly _role: iam.Role;
  private readonly _policy: iam.Policy;

  get policy(): iam.Policy {
    return this._policy;
  }

  get role(): iam.Role {
    return this._role;
  }

  constructor(scope: Construct, id: string, props: RoleProps) {
    super(scope, id);

    const healthLakeArn = Fn.importValue(
      "v2-HealthlakeInfra-primary-store-arn",
    );

    const keyArn = Fn.importValue("account-data-kmskey");

    // this._policy = new Policy(scope, "HealthLakeExportPolicy", {
    const manageExportPolicy = new iam.PolicyDocument({
      statements: [
        new iam.PolicyStatement({
          effect: iam.Effect.ALLOW,
          resources: [healthLakeArn],
          actions: [
            "healthlake:StartFHIRExportJobWithPost",
            "healthlake:DescribeFHIRExportJobWithGet",
            "healthlake:CancelFHIRExportJobWithDelete",
            "healthlake:StartFHIRExportJob",
          ],
        }),
        new iam.PolicyStatement({
          actions: [
            "s3:ListBucket",
            "s3:GetBucketPublicAccessBlock",
            "s3:GetEncryptionConfiguration",
          ],
          effect: iam.Effect.ALLOW,
          resources: [`${props.bucket.bucketArn}`],
        }),
        new iam.PolicyStatement({
          actions: ["s3:PutObject"],
          effect: iam.Effect.ALLOW,
          resources: [`${props.bucket.bucketArn}/*`],
        }),

        new iam.PolicyStatement({
          actions: ["kms:DescribeKey", "kms:GenerateDataKey*"],
          effect: iam.Effect.ALLOW,
          resources: [keyArn],
        }),
      ],
    });

    const assumedBy = new iam.ServicePrincipal("healthlake.amazonaws.com");
    this._role = new iam.Role(scope, "HealthLakeExportRole", {
      roleName: "healthlake-cdc-export-role",
      assumedBy: new iam.PrincipalWithConditions(assumedBy, {
        StringEquals: {
          "aws:SourceAccount": getAccountId(props.stage),
        },
        ArnEquals: {
          "aws:SourceArn": healthLakeArn,
        },
      }),
      inlinePolicies: {
        healthlakePolicy: manageExportPolicy,
      },
    });
  }
}
