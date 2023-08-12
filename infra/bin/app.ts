#!/usr/bin/env node
import "source-map-support/register";
import * as cdk from "aws-cdk-lib";
import { getConfig } from "./config";
import { PipelineStack } from "../lib/pipeline/pipeline-stack";

const app = new cdk.App();
const config = getConfig("main", "HealthLakeCdcExport");

// new MainStack(app, `main-HealthLakeCdcAttemptTwo`, {
//     options: config,
//     stageEnvironment: StageEnvironment.DEV,
//     tags: {
//         billingTag: "mda-pipeline",
//         service: "mda-pipeline",
//         subService: "change-data-capture",
//     },
// });

new PipelineStack(
  app,
  `${config.stackNamePrefix}-${config.stackName}-PipelineStack`,
  {
    env: {
      account: config.toolsAccount,
      region: config.defaultRegion,
    },
    options: config,
  },
);
