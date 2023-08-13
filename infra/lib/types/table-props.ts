import { IKey } from "aws-cdk-lib/aws-kms";

export interface TableProps {
    key: IKey;
}
