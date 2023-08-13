import { Stack, StackProps } from "aws-cdk-lib";
import { Construct } from "constructs";
import { StackOptions } from "./types/stack-options";
import { Effect, PolicyStatement } from "aws-cdk-lib/aws-iam";
import CdcBucket from "./constructs/cdc-bucket";
import { CdcRole } from "./constructs/cdc-role";
import QueueConstruct from "./constructs/queue-construct";
import TableConstruct from "./constructs/table-construct";
import FunctionConstruct from "./constructs/function-construct";
import CdcStateMachineConstruct from "./constructs/cdc-state-machine-construct";
import { ScheduleConstruct } from "./constructs/schedule-construct";
import KmsConstruct from "./constructs/kms-construct";
import { HealthLakeConstruct } from "./constructs/healthlake-construct";
import { EventBusConstruct } from "./constructs/event-bus-construct";

interface MainStackProps extends StackProps {
    options: StackOptions;
}

export class MainStack extends Stack {
    constructor(scope: Construct, id: string, props: MainStackProps) {
        super(scope, id, props);

        const version = new Date().toISOString();
        const accountId = Stack.of(this).account;
        const keyConstruct = new KmsConstruct(this, "KmsConstruct");
        const healthlakeConstruct = new HealthLakeConstruct(
            this,
            "HealthLakeConstruct",
            {
                key: keyConstruct.key,
            }
        );

        const bucketConstruct = new CdcBucket(this, "CdcBucketConstruct", {
            key: keyConstruct.key,
        });
        const iam = new CdcRole(this, "CdcIam", {
            bucket: bucketConstruct.bucket,
            datastore: healthlakeConstruct.datastore,
            key: keyConstruct.key,
            accountId: accountId,
        });

        const busConstruct = new EventBusConstruct(this, "EventBusConstruct");
        const q = new QueueConstruct(this, "RecheckQueueConstruct", {
            key: keyConstruct.key,
        });

        const t = new TableConstruct(this, "ExportTable", {
            key: keyConstruct.key,
        });
        const f = new FunctionConstruct(this, "FunctionsConstruct", {
            options: props.options,
            version: version,
            recheckQueue: q.reCheckQueue,
            startExportQueue: q.startExportQueue,
            bucket: bucketConstruct.bucket,
            role: iam.role,
            datastore: healthlakeConstruct.datastore,
            key: keyConstruct.key,
        });

        const cdc = new CdcStateMachineConstruct(this, "CdcExportConstruct", {
            exportTable: t.exportTable,
            stackOptions: props.options,
            startExportQueue: q.startExportQueue,
            bucket: bucketConstruct.bucket,
            prepChangeFunction: f.prepareChangeFunction,
            bus: busConstruct.bus,
            datastore: healthlakeConstruct.datastore,
            key: keyConstruct.key,
        });

        // cdc.stateMachine.grantTaskResponse(f.describeExportFunction);
        f.describeExportFunction.addToRolePolicy(
            new PolicyStatement({
                actions: [
                    "states:SendTaskFailure",
                    "states:SendTaskHeartbeat",
                    "states:SendTaskSuccess",
                ],
                effect: Effect.ALLOW,
                resources: [cdc.sf.stateMachineArn],
            })
        );

        new ScheduleConstruct(this, "ScheduleConstruct", {
            stateMachine: cdc.sf,
        });
    }
}
