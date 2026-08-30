package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type telegramUpdate struct {
	UpdateID int64 `json:"update_id"`
	Message *struct {
		Chat struct { ID int64 `json:"id"`; Type string `json:"type"` } `json:"chat"`
		Text string `json:"text"`
	} `json:"message"`
}

type telegramResponse struct {
	OK bool `json:"ok"`
	Result []telegramUpdate `json:"result"`
}

func startTelegramBot() {
	token := strings.TrimSpace(os.Getenv("BOT_TOKEN"))
	if token == "" {
		log.Println("[INFO] Telegram bot disabled: BOT_TOKEN is not configured")
		return
	}
	go pollTelegram(token)
}

func telegramAPI(token, method string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil { return err }
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Post("https://api.telegram.org/bot"+token+"/"+method, "application/json", bytes.NewReader(data))
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 { return fmt.Errorf("telegram api status %d", resp.StatusCode) }
	var result struct { OK bool `json:"ok"`; Description string `json:"description"` }
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil { return err }
	if !result.OK { return fmt.Errorf("telegram: %s", result.Description) }
	return nil
}

func sendBotMessage(token string, chatID int64, text string, keyboard any) {
	payload := map[string]any{"chat_id": chatID, "text": text, "disable_web_page_preview": true}
	if keyboard != nil { payload["reply_markup"] = keyboard }
	if err := telegramAPI(token, "sendMessage", payload); err != nil { log.Printf("[WARN] sendMessage: %v", err) }
}

func botKeyboard() any {
	webURL := strings.TrimSpace(os.Getenv("WEBAPP_URL"))
	if webURL == "" { return nil }
	return map[string]any{"inline_keyboard": [][]map[string]any{{{"text":"🎵 Open Music","web_app":map[string]string{"url":webURL}}}}}
}

func pollTelegram(token string) {
	_ = telegramAPI(token, "deleteWebhook", map[string]any{"drop_pending_updates": false})
	_ = telegramAPI(token, "setMyCommands", map[string]any{"commands": []map[string]string{
		{"command":"start", "description":"Start Music"},
		{"command":"help", "description":"Show help"},
		{"command":"music", "description":"Open Music"},
	}})

	var offset int64
	for {
		values := url.Values{}
		values.Set("timeout", "25")
		values.Set("offset", strconv.FormatInt(offset, 10))
		req, _ := http.NewRequest(http.MethodGet, "https://api.telegram.org/bot"+token+"/getUpdates?"+values.Encode(), nil)
		resp, err := (&http.Client{Timeout: 35 * time.Second}).Do(req)
		if err != nil { log.Printf("[WARN] Telegram polling: %v", err); time.Sleep(3*time.Second); continue }
		var result telegramResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		if err != nil || !result.OK { time.Sleep(3*time.Second); continue }

		for _, update := range result.Result {
			if update.UpdateID >= offset { offset = update.UpdateID + 1 }
			if update.Message == nil { continue }
			text := strings.TrimSpace(update.Message.Text)
			cmd := strings.ToLower(strings.Fields(text)[0])
			if text == "" { continue }
			if i := strings.Index(cmd, "@"); i >= 0 { cmd = cmd[:i] }

			switch cmd {
			case "/start":
				sendBotMessage(token, update.Message.Chat.ID, "🎵 Welcome to Music!\n\nSearch and play music from the Mini App.", botKeyboard())
			case "/help":
				sendBotMessage(token, update.Message.Chat.ID, "🎵 Music Help\n\n/start — Start the bot\n/music — Open Music Mini App\n/help — Show this help\n\nVC Cast is optional and stays disabled until its credentials are configured.", botKeyboard())
			case "/music":
				if os.Getenv("WEBAPP_URL") == "" {
					sendBotMessage(token, update.Message.Chat.ID, "Music Mini App URL is not configured yet.", nil)
				} else {
					sendBotMessage(token, update.Message.Chat.ID, "Tap below to open Music 🎵", botKeyboard())
				}
			}
		}
	}
}
