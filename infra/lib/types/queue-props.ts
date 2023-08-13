import { IKey } from "aws-cdk-lib/aws-kms";

export interface QueueProps {
    key: IKey;
}
