import { StateMachine } from "aws-cdk-lib/aws-stepfunctions";

export interface ScheduleProps {
  stateMachine: StateMachine;
}
