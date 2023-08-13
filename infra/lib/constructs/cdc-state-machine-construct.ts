import { Construct } from "constructs";
import { StateMachineConstructProps } from "../types/state-machine-props";
import {
    DefinitionBody,
    LogLevel,
    StateMachine,
    StateMachineType,
} from "aws-cdk-lib/aws-stepfunctions";
import { LogGroup, RetentionDays } from "aws-cdk-lib/aws-logs";
import { Fn, RemovalPolicy } from "aws-cdk-lib";
import * as fs from "fs";
import * as path from "path";
import {
    Effect,
    PolicyDocument,
    PolicyStatement,
    Role,
    ServicePrincipal,
} from "aws-cdk-lib/aws-iam";

export default class CdcStateMachineConstruct extends Construct {
    private readonly _sf: StateMachine;

    get sf(): StateMachine {
        return this._sf;
    }

    constructor(
        scope: Construct,
        id: string,
        props: StateMachineConstructProps
    ) {
        super(scope, id);

        const file = fs.readFileSync(
            path.resolve(__dirname, "../asl/cdc.json")
        );

        const logGroup = new LogGroup(this, "CloudwatchLogs", {
            logGroupName: "/aws/vendedlogs/states/healthlake-export",
            removalPolicy: RemovalPolicy.DESTROY,
            retention: RetentionDays.ONE_DAY,
        });

        const busPolicy = new PolicyDocument({
            statements: [
                new PolicyStatement({
                    actions: ["events:PutEvents"],
                    resources: [props.bus.eventBusArn],
                    effect: Effect.ALLOW,
                }),
            ],
        });

        const kmsPolicy = new PolicyDocument({
            statements: [
                new PolicyStatement({
                    actions: [
                        "kms:Decrypt",
                        "kms:DescribeKey",
                        "kms:Encrypt",
                        "kms:GenerateDataKey*",
                        "kms:ReEncrypt*",
                    ],
                    resources: [props.key.keyArn],
                    effect: Effect.ALLOW,
                }),
            ],
        });

        const ddbPolicy = new PolicyDocument({
            statements: [
                new PolicyStatement({
                    actions: [
                        "dynamodb:GetItem",
                        "dynamodb:BatchGetItem",
                        "dynamodb:BatchWriteItem",
                        "dynamodb:ConditionCheckItem",
                        "dynamodb:DeleteItem",
                        "dynamodb:DescribeTable",
                        "dynamodb:GetRecords",
                        "dynamodb:GetShardIterator",
                        "dynamodb:PutItem",
                        "dynamodb:Query",
                        "dynamodb:Scan",
                        "dynamodb:UpdateItem",
                    ],
                    resources: [props.exportTable.tableArn],
                    effect: Effect.ALLOW,
                }),
            ],
        });

        const logPolicy = new PolicyDocument({
            statements: [
                new PolicyStatement({
                    actions: [
                        "logs:CreateLogDelivery",
                        "logs:DeleteLogDelivery",
                        "logs:DescribeLogGroups",
                        "logs:DescribeResourcePolicies",
                        "logs:GetLogDelivery",
                        "logs:ListLogDeliveries",
                        "logs:PutResourcePolicy",
                        "logs:UpdateLogDelivery",
                    ],
                    resources: ["*"],
                    effect: Effect.ALLOW,
                }),
            ],
        });

        const queuePolicy = new PolicyDocument({
            statements: [
                new PolicyStatement({
                    actions: [
                        "sqs:GetQueueAttributes",
                        "sqs:GetQueueUrl",
                        "sqs:SendMessage",
                    ],
                    resources: [props.startExportQueue.queueArn],
                    effect: Effect.ALLOW,
                }),
            ],
        });

        const s3Policy = new PolicyDocument({
            statements: [
                new PolicyStatement({
                    actions: ["s3:*"],
                    effect: Effect.ALLOW,
                    resources: [
                        props.bucket.bucketArn,
                        `${props.bucket.bucketArn}/*`,
                    ],
                }),
            ],
        });

        const lambdaInvoke = new PolicyDocument({
            statements: [
                new PolicyStatement({
                    resources: ["*"],
                    effect: Effect.ALLOW,
                    actions: ["lambda:InvokeFunction"],
                }),
            ],
        });

        const distributeMapPolicy = new PolicyDocument({
            statements: [
                new PolicyStatement({
                    actions: [
                        "states:StartExecution",
                        "states:DescribeExecution",
                        "states:StopExecution",
                    ],
                    effect: Effect.ALLOW,
                    resources: ["*"],
                }),
            ],
        });

        const role = new Role(this, "StateMachineRole", {
            assumedBy: new ServicePrincipal(`states.us-west-2.amazonaws.com`),
            inlinePolicies: {
                s3ReadWrite: s3Policy,
                cloudwatch: logPolicy,
                ddb: ddbPolicy,
                startExportQueue: queuePolicy,
                dataAndQueueKey: kmsPolicy,
                distributedMap: distributeMapPolicy,
                lambdaInvoke: lambdaInvoke,
                eventBridgeEvents: busPolicy,
            },
        });

        this._sf = new StateMachine(this, "Cdc", {
            definitionBody: DefinitionBody.fromString(file.toString()),
            role: role,
            stateMachineName: "HealthLakeCDCExport",
            stateMachineType: StateMachineType.STANDARD,
            logs: {
                level: LogLevel.ALL,
                destination: logGroup,
                includeExecutionData: true,
            },
            definitionSubstitutions: {
                PrepareChangeFunction: props.prepChangeFunction.functionArn,
                StartExportQueueUrl: props.startExportQueue.queueUrl,
            },
        });
    }
}
