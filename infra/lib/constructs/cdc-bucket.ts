import { Duration, Fn, RemovalPolicy } from "aws-cdk-lib";
import { Key } from "aws-cdk-lib/aws-kms";
import {
    BlockPublicAccess,
    Bucket,
    BucketEncryption,
    IntelligentTieringConfiguration,
} from "aws-cdk-lib/aws-s3";
import { Construct } from "constructs";
import { BucketProps } from "../types/bucket-props";

export default class CdcBucket extends Construct {
    private readonly _bucket: Bucket;

    get bucket(): Bucket {
        return this._bucket;
    }

    constructor(scope: Construct, id: string, props: BucketProps) {
        super(scope, id);

        const authArchive: IntelligentTieringConfiguration = {
            name: "auto-archive",
            archiveAccessTierTime: Duration.days(90),
            deepArchiveAccessTierTime: Duration.days(180),
        };

        this._bucket = new Bucket(scope, "CdcBucket", {
            encryption: BucketEncryption.KMS,
            encryptionKey: props.key,
            intelligentTieringConfigurations: [authArchive],
            blockPublicAccess: BlockPublicAccess.BLOCK_ALL,
            removalPolicy: RemovalPolicy.DESTROY,
            autoDeleteObjects: true,
        });
    }
}
