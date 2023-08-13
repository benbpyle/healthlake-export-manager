import { Construct } from "constructs";
import { HealthLakeProps } from "../types/healthlake-props";
import { CfnFHIRDatastore } from "aws-cdk-lib/aws-healthlake";

export class HealthLakeConstruct extends Construct {
    private _datastore: CfnFHIRDatastore;

    get datastore(): CfnFHIRDatastore {
        return this._datastore;
    }

    constructor(scope: Construct, id: string, props: HealthLakeProps) {
        super(scope, id);

        this._datastore = new CfnFHIRDatastore(this, "HealthlakeDataStore", {
            datastoreTypeVersion: "R4",
            datastoreName: `sample-datastore`,
            sseConfiguration: {
                kmsEncryptionConfig: {
                    cmkType: "CUSTOMER_MANAGED_KMS_KEY",
                    kmsKeyId: props.key.keyId,
                },
            },
        });
    }
}
