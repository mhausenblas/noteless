NOTELESS_BUCKET?=noteless-code
NOTELESS_FRONTEND_BUCKET?=noteless-static
NOTELESS_STACK?=noteless
TARGET_REGION?=eu-west-1

.PHONY: build buildintake buildiconize up deploy frontendup showapi destroy status

build: buildintake buildiconize

buildintake:
	GOOS=linux GOARCH=amd64 go build -v -ldflags '-d -s -w' -a -tags netgo -installsuffix netgo -o bin/intake ./intake

buildiconize:
	GOOS=linux GOARCH=amd64 go build -v -ldflags '-d -s -w' -a -tags netgo -installsuffix netgo -o bin/iconize ./iconize

up: 
	sam package --template-file template.yaml --output-template-file current-stack.yaml --s3-bucket ${NOTELESS_BUCKET}
	sam deploy --template-file current-stack.yaml --stack-name ${NOTELESS_STACK} --capabilities CAPABILITY_IAM

deploy: build up

showapi: 
	@aws cloudformation describe-stacks --stack-name ${NOTELESS_STACK} --query "Stacks[0].Outputs[?OutputKey=='NotelessAPIURL'].OutputValue" --output text

destroy:
	aws cloudformation delete-stack --stack-name ${NOTELESS_STACK}

status:
	aws cloudformation describe-stacks --stack-name ${NOTELESS_STACK}