import { CfnFHIRDatastore } from "aws-cdk-lib/aws-healthlake";
import { IKey } from "aws-cdk-lib/aws-kms";
import { Bucket } from "aws-cdk-lib/aws-s3";

export interface RoleProps {
    bucket: Bucket;
    datastore: CfnFHIRDatastore;
    key: IKey;
    accountId: string;
}
