import { Stack, StackProps } from "aws-cdk-lib";
import { Construct } from "constructs";
import { StageEnvironment } from "./types/stage-environment";
import { StackOptions } from "./types/stack-options";
import { Effect, PolicyStatement } from "aws-cdk-lib/aws-iam";
import CdcBucket from "./constructs/cdc-bucket";
import { CdcRole } from "./constructs/cdc-role";
import QueueConstruct from "./constructs/queue-construct";
import TableConstruct from "./constructs/table-construct";
import FunctionConstruct from "./constructs/function-construct";
import CdcStateMachineConstruct from "./constructs/cdc-state-machine-construct";
import { ScheduleConstruct } from "./constructs/schedule-construct";

interface MainStackProps extends StackProps {
  options: StackOptions;
  stageEnvironment: StageEnvironment;
}

export class MainStack extends Stack {
  constructor(scope: Construct, id: string, props: MainStackProps) {
    super(scope, id, props);

    const version = new Date().toISOString();

    const bucketConstruct = new CdcBucket(this, "CdcBucketConstruct", {
      stage: props.stageEnvironment,
    });
    const iam = new CdcRole(this, "CdcIam", {
      bucket: bucketConstruct.bucket,
      stage: props.stageEnvironment,
    });

    const q = new QueueConstruct(this, "RecheckQueueConstruct", props.options);

    const t = new TableConstruct(this, "ExportTable");
    const f = new FunctionConstruct(this, "FunctionsConstruct", {
      options: props.options,
      stage: props.stageEnvironment,
      version: version,
      queueKey: q.queueKey,
      recheckQueue: q.reCheckQueue,
      startExportQueue: q.startExportQueue,
      bucket: bucketConstruct.bucket,
      role: iam.role,
    });

    const cdc = new CdcStateMachineConstruct(this, "CdcExportConstruct", {
      dataKey: t.dataKey,
      exportTable: t.exportTable,
      stackOptions: props.options,
      startExportQueue: q.startExportQueue,
      queueKey: q.queueKey,
      bucket: bucketConstruct.bucket,
      prepChangeFunction: f.prepareChangeFunction,
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
      }),
    );

    new ScheduleConstruct(this, "ScheduleConstruct", {
      stateMachine: cdc.sf,
    });
  }
}
