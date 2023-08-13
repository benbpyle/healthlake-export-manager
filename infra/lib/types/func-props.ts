import { Queue } from "aws-cdk-lib/aws-sqs";
import { StackOptions } from "./stack-options";
import { IKey } from "aws-cdk-lib/aws-kms";
import { Bucket } from "aws-cdk-lib/aws-s3";
import { Role } from "aws-cdk-lib/aws-iam";
import { CfnFHIRDatastore } from "aws-cdk-lib/aws-healthlake";

export interface FuncProps {
    options: StackOptions;
    version: string;
    startExportQueue: Queue;
    recheckQueue: Queue;
    bucket: Bucket;
    role: Role;
    key: IKey;
    datastore: CfnFHIRDatastore;
}
