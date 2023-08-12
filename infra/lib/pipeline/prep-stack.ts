import { Stack, StackProps } from "aws-cdk-lib";
import { Construct } from "constructs";
import { CfnEventBusPolicy, CfnRule } from "aws-cdk-lib/aws-events";
import { StackOptions } from "../types/stack-options";

interface PipelinePrepStackProps extends StackProps {
  options: StackOptions;
}

export class PipelinePrepStack extends Stack {
  /**
   * Creates the Event Bus Policy required by the Pipeline in the Tools account.
   * Allows for triggering of the Pipeline from CodeCommit in a different Account.
   * Only required for cross-account CodeCommit/Pipeline.
   *
   * @param {Construct} scope
   * @param {string} id
   * @param {StackProps=} props
   */
  constructor(scope: Construct, id: string, props: PipelinePrepStackProps) {
    super(scope, id, props);

    const { options } = props;

    // Allow CodeCommit account EventBus to put events to Pipeline account EventBus
    // This is used to trigger the pipeline from CodeCommit updates in the Development account
    new CfnEventBusPolicy(
      this,
      `${props?.options.stackNamePrefix}-${props?.options.stackName}-EventsPolicy`,
      {
        statementId: `${props?.options.stackNamePrefix}-${props?.options.stackName}-CodeCommit`,
        eventBusName: "default",
        statement: {
          Effect: "Allow",
          Principal: {
            AWS: `arn:aws:iam::${props.options.codeCommitAccount}:root`,
          },
          Action: "events:PutEvents",
          Resource: `arn:aws:events:${this.region}:${this.account}:event-bus/default`,
        },
      },
    );

    // Base Role for pipelines. Created here as it is required outside of the pipeline stack for cross-region deployments.
  }
}

interface CodeCommitStackProps extends StackProps {
  options: StackOptions;
}

export class CodeCommitStack extends Stack {
  /**
   * Creates the CodeCommit Repository.
   *
   * @param {Construct} scope
   * @param {string} id
   * @param {StackProps=} props
   */
  constructor(scope: Construct, id: string, props: CodeCommitStackProps) {
    super(scope, id, props);

    const { options } = props;

    // Create an Events rule to send all CodeCommit repository updates for our repo to the Pipeline Account.
    // They are filtered by branch at the other end by the Pipeline rules.
    // The Event Bus Policy in the Pipeline account must be created to allow this first (in the PipelinePrepStack above).
    new CfnRule(
      this,
      `${props?.options.stackNamePrefix}-${props?.options.stackName}-UpdateToPipeline`,
      {
        description: "Send CodeCommit events to Pipeline Account",
        eventBusName: "default",
        eventPattern: {
          "detail-type": ["CodeCommit Repository State Change"],
          source: ["aws.codecommit"],
          resources: [
            `arn:aws:codecommit:${this.region}:${this.account}:${props.options.reposName}`,
          ],
        },
        state: "ENABLED",
        targets: [
          {
            arn: `arn:aws:events:${this.region}:${props.options.toolsAccount}:event-bus/default`,
            id: `${props?.options.stackNamePrefix}-${props?.options.stackName}-PipelineTarget`,
          },
        ],
      },
    );
  }
}
