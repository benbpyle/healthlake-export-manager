run:
	cdk synth
	sam local start-api -t cdk.out/main-UserProfile.template.json --env-vars environment.json --skip-pull-image
deploy-local: 
	cdk synth
	cdk deploy  `npx ts-node bin/app.ts` --profile=dev
build-run-handler:
	cdk synth
	make run-handler
run-handler:
	sam local invoke UserUpdaterFunction -t cdk.out/main-UserProfile.template.json --event src/user-dynamodb-updater/test-events/one.json --env-vars environment.json --profile=dev --skip-pull-image

