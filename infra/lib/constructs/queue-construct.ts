import { Duration, Fn } from "aws-cdk-lib";
import { IKey, Key } from "aws-cdk-lib/aws-kms";
import { Queue, QueueEncryption } from "aws-cdk-lib/aws-sqs";
import { Construct } from "constructs";
import { StackOptions } from "../types/stack-options";

export default class QueueConstruct extends Construct {
  private readonly _reCheckQueue: Queue;
  private readonly _startExportQueue: Queue;
  private readonly _reCheckDeadLetterQueue: Queue;
  private readonly _startExportDeadLetterQueue: Queue;
  private readonly _queueKey: IKey;

  constructor(scope: Construct, id: string, props: StackOptions) {
    super(scope, id);

    const sqsKeyArn = Fn.importValue("account-sns-sqs-kmskey");
    this._queueKey = Key.fromKeyArn(this, "SqsKey", sqsKeyArn);
    this._reCheckDeadLetterQueue = new Queue(
      this,
      `${props.stackNamePrefix}-${props.stackName}-Recheck-DLQ`,
      {
        queueName: `healthlake-cdc-recheck-dlq`,
        encryption: QueueEncryption.KMS,
        encryptionMasterKey: this._queueKey,
      },
    );

    this._startExportDeadLetterQueue = new Queue(
      this,
      `${props.stackNamePrefix}-${props.stackName}-StartExport-DLQ`,
      {
        queueName: `healthlake-cdc-start-export-dlq`,
        encryption: QueueEncryption.KMS,
        encryptionMasterKey: this._queueKey,
      },
    );

    this._reCheckQueue = new Queue(
      this,
      `${props.stackNamePrefix}-${props.stackName}-Recheck-Queue`,
      {
        queueName: `healthlake-cdc-recheck-queue`,
        encryption: QueueEncryption.KMS,
        encryptionMasterKey: this._queueKey,
        deadLetterQueue: {
          maxReceiveCount: 1,
          queue: this._reCheckDeadLetterQueue,
        },
      },
    );

    this._startExportQueue = new Queue(
      this,
      `${props.stackNamePrefix}-${props.stackName}-StartExport-Queue`,
      {
        queueName: `healthlake-cdc-start-export-queue`,
        encryption: QueueEncryption.KMS,
        encryptionMasterKey: this._queueKey,
        visibilityTimeout: Duration.seconds(90),
        deadLetterQueue: {
          maxReceiveCount: 1,
          queue: this._startExportDeadLetterQueue,
        },
      },
    );
  }

  get reCheckQueue(): Queue {
    return this._reCheckQueue;
  }

  get startExportQueue(): Queue {
    return this._startExportQueue;
  }

  get queueKey(): IKey {
    return this._queueKey;
  }
}
