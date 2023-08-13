import { IKey } from "aws-cdk-lib/aws-kms";

export interface HealthLakeProps {
    key: IKey;
}
