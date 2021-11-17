
build:
	sam build

deploy: build
	sam deploy
	@BOT_BASE_URL=$$(aws cloudformation describe-stacks --stack-name devch2021-bot --query 'Stacks[0].Outputs[?OutputKey==`BotBaseUrl`].OutputValue' --output text); \
	curl -f "https://api.telegram.org/bot$(TELEBOT_SECRET)/setWebhook?url="$$BOT_BASE_URL"tg-webhook"
