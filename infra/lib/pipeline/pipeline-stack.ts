import * as cdk from "aws-cdk-lib";
import { StackProps } from "aws-cdk-lib";
import { Construct } from "constructs";
import { Repository } from "aws-cdk-lib/aws-codecommit";
import {
  CodePipeline,
  CodePipelineSource,
  ShellStep,
} from "aws-cdk-lib/pipelines";
import { Effect, PolicyStatement } from "aws-cdk-lib/aws-iam";
import { PipelineAppStage } from "./pipeline-app-stage";
import { BuildSpec, LinuxBuildImage } from "aws-cdk-lib/aws-codebuild";
import { StackOptions } from "../types/stack-options";
import { StageEnvironment } from "../types/stage-environment";

interface PipelineStackProps extends StackProps {
  options: StackOptions;
}

export class PipelineStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: PipelineStackProps) {
    super(scope, id, props);

    const repos = Repository.fromRepositoryArn(
      this,
      `${props?.options.stackNamePrefix}-${props?.options.stackName}-repository`,
      `arn:aws:codecommit:${props?.options.defaultRegion}:${props?.options.codeCommitAccount}:${props?.options.reposName}`,
    );
    const pipeline = new CodePipeline(
      this,
      `${props?.options.stackNamePrefix}-${props?.options.stackName}-Pipeline`,
      {
        crossAccountKeys: true,
        selfMutation: true,
        pipelineName: `${props.options.stackNamePrefix}-${props.options.reposName}-pipeline`,
        dockerEnabledForSynth: true,
        synth: new ShellStep("Synth", {
          input: CodePipelineSource.codeCommit(repos, "main"),
          commands: ["npm ci", "npm run build", "npx cdk synth"],
        }),
        selfMutationCodeBuildDefaults: {
          rolePolicy: [
            new PolicyStatement({
              sid: "CcAccountRole",
              effect: Effect.ALLOW,
              actions: ["sts:AssumeRole"],
              resources: [
                // CodeCommit account cdk roles - to allow for update of the support stack during self mutation
                `arn:aws:iam::${props?.options.codeCommitAccount}:role/cdk-${props?.options.cdkBootstrapQualifier}-deploy-role-${props?.options.codeCommitAccount}-${this.region}`,
                `arn:aws:iam::${props?.options.codeCommitAccount}:role/cdk-${props?.options.cdkBootstrapQualifier}-file-publishing-role-${props?.options.codeCommitAccount}-${this.region}`,
              ],
            }),
          ],
        },
        synthCodeBuildDefaults: {
          buildEnvironment: {
            buildImage: LinuxBuildImage.STANDARD_7_0,
          },
          partialBuildSpec: BuildSpec.fromObject({
            phases: {
              install: {
                "runtime-versions": {
                  golang: "1.20",
                },
              },
            },
          }),
        },
      },
    );

    pipeline.addStage(
      new PipelineAppStage(
        this,
        `${props?.options.stackNamePrefix}-${props?.options.stackName}-DevDeploymentStage`,
        {
          options: props.options,
          stage: StageEnvironment.DEV,
          env: {
            account: props?.options?.devAccount,
            region: props?.options?.defaultRegion,
          },
        },
      ),
    );

    pipeline.addStage(
      new PipelineAppStage(
        this,
        `${props?.options.stackNamePrefix}-${props?.options.stackName}-QADeploymentStage`,
        {
          options: props.options,
          stage: StageEnvironment.QA,
          env: {
            account: props?.options?.qaAccount,
            region: props?.options?.defaultRegion,
          },
        },
      ),
    );

    pipeline.addStage(
      new PipelineAppStage(
        this,
        `${props?.options.stackNamePrefix}-${props?.options.stackName}-ProdDeploymentStage`,
        {
          options: props.options,
          stage: StageEnvironment.PROD,
          env: {
            account: props?.options?.productionAccount,
            region: props?.options?.defaultRegion,
          },
        },
      ),
    );
  }
}
