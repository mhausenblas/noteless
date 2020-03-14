# noteless

## Usage

The **noteless** serverless demo allows you to do two things:

- [capture](capture/) text: use your phone to capture notes in print or handwritten version on paper
- [analyse](notes/) notes: by applying rules, figure out patterns or find stuff

## Background

This serverless end-to-end demo uses:

1. [Amazon Rekognition](https://aws.amazon.com/rekognition/) for detecting text in images
2. [AWS Lambda](https://aws.amazon.com/lambda/) for the capture/frontend processing
3. [Amazon EKS](https://aws.amazon.com/eks/) on [AWS Fargate](https://aws.amazon.com/fargate/) for the event-driven analytics part with an [Open Policy Agent](https://www.openpolicyagent.org/) Rego-based set of rules.
4. [Amazon S3](https://aws.amazon.com/s3/) and [Amazon DynamoDB](https://aws.amazon.com/dynamodb/) for storing the capture images and the detected text.

Might wanna check out the [architecture](https://mhausenblas.info/noteless/docs/design.pdf) and if you want to try it out yourself, 
the source code is available via [mhausenblas/noteless](https://github.com/mhausenblas/noteless).
