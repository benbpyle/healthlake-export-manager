import { RemovalPolicy } from "aws-cdk-lib";
import { IKey, Key } from "aws-cdk-lib/aws-kms";
import { Construct } from "constructs";

export default class KmsConstruct extends Construct {
    private readonly _key: IKey;
    get key(): IKey {
        return this._key;
    }

    constructor(scope: Construct, id: string) {
        super(scope, id);

        this._key = new Key(scope, "MainKey", {
            removalPolicy: RemovalPolicy.DESTROY,
        });
    }
}
