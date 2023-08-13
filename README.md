# HealthLake Export Manager

Purpose: Working example of using AWS Step Functions, SQS, Lambda, S3 and EventBridge to manage scheduled exports of FHIR resources in AWS HealthLake.

## Getting Started

### Deploying

First off, install [Node.js](https://nodejs.org/en)

```bash
# install AWS CDK
npm install -g aws-cdk
# clone the repository
cd healthlake-export-manager
npm install
```

Once dependencies have been installed, you are ready to run CDK

```bash
cdk deploy
```

## Destroying

Simply run:

```bash
cdk destroy
```

## Implementation

For a further and in-depth review of how to use this repository and what it supports, head on over the [Blog Article](https://binaryheap.com) <Article not complete but under construction>

## State Machine

![State Machine](./workflow.svg)
