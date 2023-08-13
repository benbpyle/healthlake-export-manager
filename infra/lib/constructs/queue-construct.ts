import { Duration } from "aws-cdk-lib";
import { Queue, QueueEncryption } from "aws-cdk-lib/aws-sqs";
import { Construct } from "constructs";
import { QueueProps } from "../types/queue-props";

export default class QueueConstruct extends Construct {
    private readonly _reCheckQueue: Queue;
    private readonly _startExportQueue: Queue;
    private readonly _reCheckDeadLetterQueue: Queue;
    private readonly _startExportDeadLetterQueue: Queue;

    constructor(scope: Construct, id: string, props: QueueProps) {
        super(scope, id);

        this._reCheckDeadLetterQueue = new Queue(this, `Recheck-DLQ`, {
            queueName: `healthlake-cdc-recheck-dlq`,
            encryption: QueueEncryption.KMS,
            encryptionMasterKey: props.key,
        });

        this._startExportDeadLetterQueue = new Queue(this, `StartExport-DLQ`, {
            queueName: `healthlake-cdc-start-export-dlq`,
            encryption: QueueEncryption.KMS,
            encryptionMasterKey: props.key,
        });

        this._reCheckQueue = new Queue(this, `Recheck-Queue`, {
            queueName: `healthlake-cdc-recheck-queue`,
            encryption: QueueEncryption.KMS,
            encryptionMasterKey: props.key,
            deadLetterQueue: {
                maxReceiveCount: 1,
                queue: this._reCheckDeadLetterQueue,
            },
        });

        this._startExportQueue = new Queue(this, `StartExport-Queue`, {
            queueName: `healthlake-cdc-start-export-queue`,
            encryption: QueueEncryption.KMS,
            encryptionMasterKey: props.key,
            visibilityTimeout: Duration.seconds(90),
            deadLetterQueue: {
                maxReceiveCount: 1,
                queue: this._startExportDeadLetterQueue,
            },
        });
    }

    get reCheckQueue(): Queue {
        return this._reCheckQueue;
    }

    get startExportQueue(): Queue {
        return this._startExportQueue;
    }
}
