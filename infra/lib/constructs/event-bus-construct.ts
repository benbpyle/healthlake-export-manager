import { EventBus } from "aws-cdk-lib/aws-events";
import { Construct } from "constructs";

export class EventBusConstruct extends Construct {
    private readonly _bus: EventBus;
    get bus(): EventBus {
        return this._bus;
    }

    constructor(scope: Construct, id: string) {
        super(scope, id);

        this._bus = new EventBus(this, "HealthLakeEventBus", {
            eventBusName: `healthlake-bus`,
        });
    }
}
