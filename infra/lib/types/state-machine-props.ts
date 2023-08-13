import { Table } from "aws-cdk-lib/aws-dynamodb";
import { StackOptions } from "./stack-options";
import { IKey } from "aws-cdk-lib/aws-kms";
import { Queue } from "aws-cdk-lib/aws-sqs";
import { Bucket } from "aws-cdk-lib/aws-s3";
import { IFunction } from "aws-cdk-lib/aws-lambda";
import { EventBus } from "aws-cdk-lib/aws-events";
import { CfnFHIRDatastore } from "aws-cdk-lib/aws-healthlake";

export interface StateMachineConstructProps {
    stackOptions: StackOptions;
    exportTable: Table;
    key: IKey;
    startExportQueue: Queue;
    bucket: Bucket;
    prepChangeFunction: IFunction;
    bus: EventBus;
    datastore: CfnFHIRDatastore;
}
