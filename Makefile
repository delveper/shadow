include .env
export

# ngrok
PUBLIC_URL := $(shell curl -s localhost:4040/api/tunnels | jq -r '.tunnels[0].public_url')

ngrok-install:
	curl -s https://ngrok-agent.s3.amazonaws.com/ngrok.asc | sudo tee /etc/apt/trusted.gpg.d/ngrok.asc >/dev/null && echo "deb https://ngrok-agent.s3.amazonaws.com buster main" | sudo tee /etc/apt/sources.list.d/ngrok.list && sudo apt update && sudo apt install ngrok
ngrok-config:
	ngrok config add-authtoken ${NGROK_TOKEN}
ngrok-expose:
	ngrok http ${PORT}

# telegram
TELEGRAM_URL := "https://api.telegram.org/bot${TELEGRAM_TOKEN}"

telegram-webhook:
	curl -F "url=${PUBLIC_URL}" "${TELEGRAM_URL}/setWebhook"