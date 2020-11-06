all: build

build: fmt vet
	go build -asmflags -trimpath

dist:
	zip -u letsencrypt-lambda.zip letsencrypt-lambda

fmt:
	go fmt $(shell find -name \*.go |xargs dirname|sort -u)

vet:
	go vet $(shell find -name \*.go |xargs dirname|sort -u)

run:
	#EMAIL=me@example.com DOMAINS=example.com,*.example.com AWS_HOSTED_ZONE_ID=Z123ABC456DEF7 ISSUER_PASSPHRASE=<secure_issuer_passphrase> CLIENT_PASSPHRASE=<secure_client_passphrase> time -p ./letsencrypt-lambda -local
	#EMAIL=me@example.com DOMAINS=example.com,*.example.com AWS_HOSTED_ZONE_ID=Z123ABC456DEF7 ISSUER_PASSPHRASE= CLIENT_PASSPHRASE= CLIENT_PASSPHRASE_SECRET_ARN=arn:aws:secretsmanager:<region>:<account-id>:secret:letsencrypt-lambda-client_passphrase-abcdef ISSUER_PASSPHRASE_SECRET_ARN=arn:aws:secretsmanager:<region>:<account-id>:secret:letsencrypt-lambda-issuer_passphrase-123456 time -p ./letsencrypt-lambda -local
	./run.sh
