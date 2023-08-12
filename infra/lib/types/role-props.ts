import { Bucket } from "aws-cdk-lib/aws-s3";
import { StageEnvironment } from "./stage-environment";

export interface RoleProps {
  bucket: Bucket;
  stage: StageEnvironment;
}
