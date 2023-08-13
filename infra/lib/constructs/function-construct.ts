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

    buildDescribeExport = (scope: Construct, props: FuncProps) => {
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
                    HEALTHLAKE_ENDPOINT:
                        "healthlake.us-west-2.amazonaws.com/datastore",
                    HEALTHLAKE_DATASTOREID: props.datastore.attrDatastoreId,
                    HEALTHLAKE_REGION: "us-west-2",
                    LOG_LEVEL: "DEBUG",
                    RECHECK_QUEUE_URL: props.recheckQueue.queueUrl,
                    BUCKET: props.bucket.bucketName,
                },
            }
        );

        props.recheckQueue.grantConsumeMessages(this._describeExportFunction);
        props.recheckQueue.grantSendMessages(this._describeExportFunction);
        this._describeExportFunction.addEventSource(
            new SqsEventSource(props.recheckQueue, {
                batchSize: 1,
                maxConcurrency: 2,
                enabled: true,
            })
        );

        props.key.grantEncryptDecrypt(this._describeExportFunction);
        this._describeExportFunction.addToRolePolicy(
            new PolicyStatement({
                actions: ["healthlake:*"],
                effect: Effect.ALLOW,
                resources: [props.datastore.attrDatastoreArn],
            })
        );

        props.bucket.grantReadWrite(this._describeExportFunction);
    };
    buildStartExport = (scope: Construct, props: FuncProps) => {
        this._startExportFunction = new GoFunction(
            scope,
            `StartExportFunction`,
            {
                entry: "src/start-export",
                functionName: `healthlake-cdc-start-export`,
                timeout: Duration.seconds(15),
                environment: {
                    DD_FLUSH_TO_LOG: "true",
                    DD_TRACE_ENABLED: "true",
                    IS_LOCAL: "false",
                    KEY_ARN: props.key.keyArn,
                    ROLE_ARN: props.role.roleArn,
                    S3_URI: `s3://${props.bucket.bucketName}/exports`,
                    HEALTHLAKE_ENDPOINT:
                        "healthlake.us-west-2.amazonaws.com/datastore",
                    HEALTHLAKE_DATASTOREID: props.datastore.attrDatastoreId,
                    HEALTHLAKE_REGION: "us-west-2",
                    LOG_LEVEL: "DEBUG",
                    RECHECK_QUEUE_URL: props.recheckQueue.queueUrl,
                    BUCKET: props.bucket.bucketName,
                },
            }
        );

        props.key.grantEncryptDecrypt(this._startExportFunction);
        props.recheckQueue.grantSendMessages(this._startExportFunction);
        this._startExportFunction.addEventSource(
            new SqsEventSource(props.startExportQueue, {
                batchSize: 1,
                enabled: true,
                maxConcurrency: 2,
            })
        );

        this._startExportFunction.addToRolePolicy(
            new PolicyStatement({
                actions: ["healthlake:*"],
                effect: Effect.ALLOW,
                resources: [props.datastore.attrDatastoreArn],
            })
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
            })
        );
    };

    buildPrepareChangeExport = (scope: Construct, props: FuncProps) => {
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
                    LOG_LEVEL: "DEBUG",
                },
            }
        );

        props.key.grantEncryptDecrypt(this._prepareChangeFunction);
        props.bucket.grantReadWrite(this._prepareChangeFunction);
    };

    constructor(scope: Construct, id: string, props: FuncProps) {
        super(scope, id);

        this.buildStartExport(scope, props);
        this.buildDescribeExport(scope, props);
        this.buildPrepareChangeExport(scope, props);
    }
}
