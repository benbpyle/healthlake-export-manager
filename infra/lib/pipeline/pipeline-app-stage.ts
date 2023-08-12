import * as cdk from "aws-cdk-lib";
import { Construct } from "constructs";
import { StackOptions } from "../types/stack-options";
import { StageEnvironment } from "../types/stage-environment";
import { MainStack } from "../main-stack";

interface PipelineStageProps extends cdk.StageProps {
  options: StackOptions;
  stage: StageEnvironment;
}

export class PipelineAppStage extends cdk.Stage {
  constructor(scope: Construct, id: string, props: PipelineStageProps) {
    super(scope, id, props);

    new MainStack(
      this,
      `${props?.options.stackNamePrefix}-${props?.options.stackName}-AppStack`,
      {
        options: props.options,
        stageEnvironment: props.stage,
        tags: {
          billingTag: "mda-pipeline",
          service: "mda-pipeline",
          subService: "healthlake-cdc",
        },
      },
    );
  }
}
