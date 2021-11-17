
build-TGWebhookFunction:
	@go test .
	CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR) -trimpath \
                  -ldflags='-w -s -X main.TelebotSecret=$(TELEBOT_SECRET) -X main.DeployURL=$(DEPLOY_URL)' \
                   *.go
