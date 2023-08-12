import { GoFunction } from "@aws-cdk/aws-lambda-go-alpha";
import { Duration, Fn } from "aws-cdk-lib";
import { IFunction } from "aws-cdk-lib/aws-lambda";
import { Construct } from "constructs";
import { getLogLevel } from "../utils/log-utils";
import { FuncProps } from "../types/func-props";
import { SqsEventSource } from "aws-cdk-lib/aws-lambda-event-sources";
import { Effect, PolicyStatement } from "aws-cdk-lib/aws-iam";
import { IKey, Key } from "aws-cdk-lib/aws-kms";

export default class FunctionConstruct extends Construct {
  private _startExportFunction: IFunction;
  private _describeExportFunction: IFunction;
  private _prepareChangeFunction: IFunction;

  get startExportFunction(): IFunction {
    return this._startExportFunction;
  }

  get describeExportFunction(): IFunction {
    return this._describeExportFunction;
  }

  get prepareChangeFunction(): IFunction {
    return this._prepareChangeFunction;
  }

  buildDescribeExport = (
    scope: Construct,
    datastoreId: string,
    healthLakeArn: string,
    keyArn: string,
    key: IKey,
    props: FuncProps,
  ) => {
    this._describeExportFunction = new GoFunction(
      scope,
      "DescribeExportFunction",
      {
        entry: "src/describe-export",
        functionName: `healthlake-cdc-describe-export`,
        timeout: Duration.seconds(15),
        environment: {
          DD_FLUSH_TO_LOG: "true",
          DD_TRACE_ENABLED: "true",
          IS_LOCAL: "false",
          HEALTHLAKE_ENDPOINT: "healthlake.us-west-2.amazonaws.com/datastore",
          HEALTHLAKE_DATASTOREID: datastoreId,
          HEALTHLAKE_REGION: "us-west-2",
          LOG_LEVEL: getLogLevel(props.stage),
          RECHECK_QUEUE_URL: props.recheckQueue.queueUrl,
          BUCKET: props.bucket.bucketName,
        },
      },
    );

    props.queueKey.grantEncryptDecrypt(this._describeExportFunction);
    props.recheckQueue.grantConsumeMessages(this._describeExportFunction);
    props.recheckQueue.grantSendMessages(this._describeExportFunction);
    this._describeExportFunction.addEventSource(
      new SqsEventSource(props.recheckQueue, {
        batchSize: 1,
        maxConcurrency: 2,
        enabled: true,
      }),
    );

    key.grantEncryptDecrypt(this._describeExportFunction);
    this._describeExportFunction.addToRolePolicy(
      new PolicyStatement({
        actions: ["healthlake:*"],
        effect: Effect.ALLOW,
        resources: [healthLakeArn],
      }),
    );

    props.bucket.grantReadWrite(this._describeExportFunction);
  };
  buildStartExport = (
    scope: Construct,
    datastoreId: string,
    healthLakeArn: string,
    keyArn: string,
    key: IKey,
    props: FuncProps,
  ) => {
    this._startExportFunction = new GoFunction(scope, `StartExportFunction`, {
      entry: "src/start-export",
      functionName: `healthlake-cdc-start-export`,
      timeout: Duration.seconds(15),
      environment: {
        DD_FLUSH_TO_LOG: "true",
        DD_TRACE_ENABLED: "true",
        IS_LOCAL: "false",
        KEY_ARN: keyArn,
        ROLE_ARN: props.role.roleArn,
        S3_URI: `s3://${props.bucket.bucketName}/exports`,
        HEALTHLAKE_ENDPOINT: "healthlake.us-west-2.amazonaws.com/datastore",
        HEALTHLAKE_DATASTOREID: datastoreId,
        HEALTHLAKE_REGION: "us-west-2",
        LOG_LEVEL: getLogLevel(props.stage),
        RECHECK_QUEUE_URL: props.recheckQueue.queueUrl,
        BUCKET: props.bucket.bucketName,
      },
    });

    props.queueKey.grantEncryptDecrypt(this._startExportFunction);
    props.recheckQueue.grantSendMessages(this._startExportFunction);
    this._startExportFunction.addEventSource(
      new SqsEventSource(props.startExportQueue, {
        batchSize: 1,
        enabled: true,
        maxConcurrency: 2,
      }),
    );

    this._startExportFunction.addToRolePolicy(
      new PolicyStatement({
        actions: ["healthlake:*"],
        effect: Effect.ALLOW,
        resources: [healthLakeArn],
      }),
    );

    this._startExportFunction.addToRolePolicy(
      new PolicyStatement({
        effect: Effect.ALLOW,
        actions: ["iam:PassRole"],
        resources: ["*"],
        conditions: {
          StringEquals: {
            "iam:PassedToService": "healthlake.amazonaws.com",
          },
        },
      }),
    );

    key.grantEncryptDecrypt(this._startExportFunction);
  };
  buildPrepareChangeExport = (
    scope: Construct,
    key: IKey,
    props: FuncProps,
  ) => {
    this._prepareChangeFunction = new GoFunction(
      scope,
      "PrepareChangeFunction",
      {
        entry: "src/prepare-change-event",
        functionName: `healthlake-cdc-prepare-change-event`,
        timeout: Duration.seconds(15),
        environment: {
          BUCKET: props.bucket.bucketName,
          DD_FLUSH_TO_LOG: "true",
          DD_TRACE_ENABLED: "true",
          IS_LOCAL: "false",
          LOG_LEVEL: getLogLevel(props.stage),
        },
      },
    );

    key.grantEncryptDecrypt(this._prepareChangeFunction);
    props.bucket.grantReadWrite(this._prepareChangeFunction);
  };

  constructor(scope: Construct, id: string, props: FuncProps) {
    super(scope, id);
    const datastoreId = Fn.importValue("v2-HealthlakeInfra-primary-store-id");
    const healthLakeArn = Fn.importValue(
      "v2-HealthlakeInfra-primary-store-arn",
    );

    const keyArn = Fn.importValue("account-data-kmskey");
    const primaryStoreKey = Key.fromKeyArn(
      scope,
      `${props?.options.stackNamePrefix}-${props?.options.stackName}-store-key`,
      keyArn,
    );

    this.buildStartExport(
      scope,
      datastoreId,
      healthLakeArn,
      keyArn,
      primaryStoreKey,
      props,
    );

    this.buildDescribeExport(
      scope,
      datastoreId,
      healthLakeArn,
      keyArn,
      primaryStoreKey,
      props,
    );
    this.buildPrepareChangeExport(scope, primaryStoreKey, props);
  }
}
