# Telegram Speech-to-Text Bot


## Useful materials

- [Build a Telegram Bot in Go in 9 minutes](https://medium.com/swlh/build-a-telegram-bot-in-go-in-9-minutes-e06ad38acef1)
- [Making an interactive Telegram bot in Go (Golang)](https://www.sohamkamani.com/golang/telegram-bot/)
- [Speech to Text](https://platform.openai.com/docs/guides/speech-to-text)


```shell
curl https://api.openai.com/v1/audio/transcriptions \
-H "Authorization: Bearer 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" \
-H "Content-Type: multipart/form-data" \
-F file="@./audio.mp3" \
-F model="whisper-1"
```

```shell

```


```shell
curl https://api.openai.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer  ${OPENAI_TOKEN}" \
  -d '{
    "model": "${MODEL}",
    "messages": [{"role": "system", "content": "You are a helpful tutor who can help me improve my English. You are kindly fix my errors if there is any and teach some grammar if needed."}, {"role":"user", "content": "Analyze my English: You are a helpful tutor that help me improve English"}]
  }' \
  | jq -r '.choices[0].message.content'
```
