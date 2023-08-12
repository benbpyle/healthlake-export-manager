import { Table } from "aws-cdk-lib/aws-dynamodb";
import { StackOptions } from "./stack-options";
import { IKey } from "aws-cdk-lib/aws-kms";
import { Queue } from "aws-cdk-lib/aws-sqs";
import { Bucket } from "aws-cdk-lib/aws-s3";
import { IFunction } from "aws-cdk-lib/aws-lambda";

export interface StateMachineConstructProps {
  stackOptions: StackOptions;
  exportTable: Table;
  dataKey: IKey;
  startExportQueue: Queue;
  queueKey: IKey;
  bucket: Bucket;
  prepChangeFunction: IFunction;
}
