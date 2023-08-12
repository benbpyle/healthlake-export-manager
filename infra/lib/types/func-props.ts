import { Queue } from "aws-cdk-lib/aws-sqs";
import { StackOptions } from "./stack-options";
import { StageEnvironment } from "./stage-environment";
import { IKey } from "aws-cdk-lib/aws-kms";
import { Bucket } from "aws-cdk-lib/aws-s3";
import { Role } from "aws-cdk-lib/aws-iam";

export interface FuncProps {
  options: StackOptions;
  stage: StageEnvironment;
  version: string;
  startExportQueue: Queue;
  recheckQueue: Queue;
  queueKey: IKey;
  bucket: Bucket;
  role: Role;
}
