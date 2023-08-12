import {
  CodePipeline,
  CodePipelineSource,
  ShellStep,
} from "aws-cdk-lib/pipelines";
import { Effect, PolicyStatement } from "aws-cdk-lib/aws-iam";
import { IRepository } from "aws-cdk-lib/aws-codecommit";
import { Construct } from "constructs";
import { StageEnvironment } from "../types/stage-environment";
import { StackOptions } from "../types/stack-options";

interface BucketConfig {
  keyArn: string;
  bucket: string;
}

export const getLogLevel = (stage: StageEnvironment): string => {
  switch (stage) {
    case StageEnvironment.DEV:
    case StageEnvironment.QA:
      return "debug";
  }

  return "error";
};

export const getBucketConfig = (stage: StageEnvironment): BucketConfig => {
  if (stage === StageEnvironment.QA) {
    return {
      bucket: "arn:aws:s3:::qa-dms-landing-curantis",
      keyArn:
        "arn:aws:kms:us-west-2:904442064295:key/1f1392d8-76e2-4187-a518-c9fd764c3f5b",
    };
  } else if (stage === StageEnvironment.PROD) {
    return {
      bucket: "arn:aws:s3:::prod-dms-landing-curantis",
      keyArn:
        "arn:aws:kms:us-west-2:904442064295:key/43ad8c42-da72-4edf-b8b8-3b91dc06bb17",
    };
  }
  return {
    bucket: "arn:aws:s3:::dev-dms-landing-curantis",
    keyArn:
      "arn:aws:kms:us-west-2:904442064295:key/362de9f7-08ce-415c-92c7-c5f903ea7724",
  };
};

export const getPipeline = (
  scope: Construct,
  repos: IRepository,
  options: StackOptions,
  branch: string,
  region: string,
  commands: string[],
): CodePipeline => {
  return new CodePipeline(scope, `CodePipeline`, {
    crossAccountKeys: true,
    selfMutation: true,
    pipelineName: `${options.pipelineName}`,
    dockerEnabledForSynth: true,
    synth: new ShellStep("Synth", {
      input: CodePipelineSource.codeCommit(repos, "main"),
      commands: commands,
    }),
    selfMutationCodeBuildDefaults: {
      rolePolicy: [
        new PolicyStatement({
          sid: "CcAccountRole",
          effect: Effect.ALLOW,
          actions: ["sts:AssumeRole"],
          resources: [
            // CodeCommit account cdk roles - to allow for update of the support stack during self mutation
            `arn:aws:iam::${options.codeCommitAccount}:role/cdk-${options.cdkBootstrapQualifier}-deploy-role-${options.codeCommitAccount}-${region}`,
            `arn:aws:iam::${options.codeCommitAccount}:role/cdk-${options.cdkBootstrapQualifier}-file-publishing-role-${options.codeCommitAccount}-${region}`,
          ],
        }),
      ],
    },
  });
};
