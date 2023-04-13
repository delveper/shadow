include .env
export

install-deps:
	sudo apt-get install -y jq \
	sudo apt-get install -y ffmpeg

run:
	go run ./cmd/main.go

# ngrok
PUBLIC_URL := $(shell curl -s localhost:4040/api/tunnels | jq -r '.tunnels[0].public_url')

server-install:
	curl -s https://ngrok-agent.s3.amazonaws.com/ngrok.asc \
	| sudo tee /etc/apt/trusted.gpg.d/ngrok.asc >/dev/null \
	&& echo "deb https://ngrok-agent.s3.amazonaws.com buster main" \
	| sudo tee /etc/apt/sources.list.d/ngrok.list \
	&& sudo apt update && sudo apt install ngrok

server-config:
	ngrok config add-authtoken ${NGROK_TOKEN}

server-expose:
	ngrok http ${PORT}

# telegram
TELEGRAM_URL := https://api.telegram.org
FILE_ID := AwACAgIAAxkBAAIB5GQ3mGUyRJpZUHwAAXzZCbs5CubfVgAC4ykAAt4quEnry0Kf4b5L8i8E
telegram-setWebhook:
	curl ${TELEGRAM_URL}/bot${TELEGRAM_TOKEN}/setWebhook?url=${PUBLIC_URL}

telegram-getFilePath:
	curl ${TELEGRAM_URL}/bot${TELEGRAM_TOKEN}/getFile?file_id=${FILE_ID} \
  | jq -r '.result.file_path'

# OpenAI
FILE_NAME := output.mp3

openai-models:
	curl -G https://api.openai.com/v1/models \
    -H "Authorization: Bearer ${OPENAI_TOKEN}" \
    -H "Content-Type: application/json"

openai-usage:
	curl -G https://api.openai.com/v1/usage \
         -H "Authorization: Bearer ${OPENAI_TOKEN}" \
         --data-urlencode "date=$$(date +%Y-%m-%d)"

openai-completions:
	curl https://api.openai.com/v1/chat/completions \
	-H "Content-Type: application/json" \
	-H "Authorization: Bearer ${OPENAI_TOKEN}" \
  	-d '{ "model": "gpt-3.5-turbo", "messages": [{"role": "system", "content": "You are a helpful tutor who can help me improve my English. You can kindly fix my errors if there are any and teach me some grammar if needed."}, {"role":"user", "content": "Analyze my English: You are a helpful tutor that helps me improve my English."}] }' \
	| jq -r '.choices[0].message.content'

openai-transcription:
	curl https://api.openai.com/v1/audio/transcriptions \
      -H "Authorization: Bearer  ${OPENAI_TOKEN}" \
      -H "Content-Type: multipart/form-data" \
      -F file="@${FILE_NAME}" \
      -F model="whisper-1" \
      | jq -r '.text'
