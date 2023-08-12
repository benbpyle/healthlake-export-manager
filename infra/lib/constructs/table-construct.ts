import { Construct } from "constructs";
import { Fn, RemovalPolicy } from "aws-cdk-lib";
import { IKey, Key } from "aws-cdk-lib/aws-kms";
import {
  AttributeType,
  BillingMode,
  Table,
  TableEncryption,
} from "aws-cdk-lib/aws-dynamodb";

export default class TableConstruct extends Construct {
  private readonly _table: Table;
  private readonly _dataKey: IKey;

  constructor(scope: Construct, id: string) {
    super(scope, id);

    const dataKeyArn = Fn.importValue("account-data-kmskey");
    this._dataKey = Key.fromKeyArn(this, "DataKey", dataKeyArn);
    this._table = new Table(this, "SinceTable", {
      billingMode: BillingMode.PAY_PER_REQUEST,
      removalPolicy: RemovalPolicy.DESTROY,
      partitionKey: { name: "id", type: AttributeType.STRING },
      pointInTimeRecovery: true,
      tableName: "HealthLakeCdcExports",
      encryption: TableEncryption.CUSTOMER_MANAGED,
      encryptionKey: this._dataKey,
    });

    /*
            {
                id: "RUN":,
                lastRunTime: "<time epoch>",
                status: "RUNNING|COMPLETED|FAILED|STOPPED"
            },
            {
                id: "RUN:<time epoch>",
                status: "COMPLETED|FAILED|STOPPED",
                s3Configuration: {
                    s3Uri: "<uri>",
                    kmsKeyId: "<kmsKeyId>"
                }
            }
        */
  }

  get exportTable(): Table {
    return this._table;
  }

  get dataKey(): IKey {
    return this._dataKey;
  }
}
