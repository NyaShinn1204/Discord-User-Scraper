package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var token string
var serverID string
var channelID string

func main() {
	args := os.Args[1:]

	if len(args) != 3 {
		fmt.Println("Usage: go run main.go <serverid> <channelid> <token>")
		return
	}

	serverID := args[0]
	channelID := args[1]
	token := args[2]

	fmt.Printf("Server ID: %s\n", serverID)
	fmt.Printf("Channel ID: %s\n", channelID)
	fmt.Printf("Token: %s\n", token)
	response := checktoken(token)
	if response == 200 {
		fmt.Printf("[VALID]   %v [%v]\n", token, response)
	} else {
		fmt.Printf("[INVALID] %v [%v]", token, response)
		os.Exit(1)
	}
	fmt.Printf("Scraping in %s with %s\n", serverID, token)
	members := getMembers(token, serverID, channelID)
	fmt.Printf("Total Scrapped: %d\n", len(members))
	userIDs := make([]string, len(members))
	for i, member := range members {
		userIDs[i] = member
	}

	fmt.Printf("Scrapped Id: %v\n", userIDs)

}

func checktoken(token string) int {
	url := "https://discord.com/api/v9/users/@me/affinities/guilds"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("%v", err)
		return -1
	}
	req.Header.Set("authorization", token)
	httpClient := &http.Client{}
	resp, err := httpClient.Do(commonheaders(req))
	if err != nil {
		fmt.Printf("%v", err)
		return -1
	}
	return resp.StatusCode
}

func commonheaders(req *http.Request) *http.Request {
	req.Header.Set("x-super-properties", "eyJvcyI6IldpbmRvd3MiLCJicm93c2VyIjoiRmlyZWZveCIsImRldmljZSI6IiIsInN5c3RlbV9sb2NhbGUiOiJlbi1VUyIsImJyb3dzZXJfdXNlcl9hZ2VudCI6Ik1vemlsbGEvNS4wIChXaW5kb3dzIE5UIDEwLjA7IFdpbjY0OyB4NjQ7IHJ2OjkzLjApIEdlY2tvLzIwMTAwMTAxIEZpcmVmb3gvOTMuMCIsImJyb3dzZXJfdmVyc2lvbiI6IjkzLjAiLCJvc192ZXJzaW9uIjoiMTAiLCJyZWZlcnJlciI6IiIsInJlZmVycmluZ19kb21haW4iOiIiLCJyZWZlcnJlcl9jdXJyZW50IjoiIiwicmVmZXJyaW5nX2RvbWFpbl9jdXJyZW50IjoiIiwicmVsZWFzZV9jaGFubmVsIjoic3RhYmxlIiwiY2xpZW50X2J1aWxkX251bWJlciI6MTAwODA0LCJjbGllbnRfZXZlbnRfc291cmNlIjpudWxsfQ==")
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("x-context-properties", "eyJsb2NhdGlvbiI6IkpvaW4gR3VpbGQiLCJsb2NhdGlvbl9ndWlsZF9pZCI6Ijg4NTkwNzE3MjMwNTgwOTUxOSIsImxvY2F0aW9uX2NoYW5uZWxfaWQiOiI4ODU5MDcxNzIzMDU4MDk1MjUiLCJsb2NhdGlvbl9jaGFubmVsX3R5cGUiOjB9")
	req.Header.Set("sec-ch-ua", "'Chromium';v='92', ' Not A;Brand';v='99', 'Google Chrome';v='92'")
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en-GB")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", "\"Windows\"")
	req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) discord/0.0.16 Chrome/91.0.4472.164 Electron/13.4.0 Safari/537.36")
	return req
}

func getMembers(token, server, channel string) []string {
	conn, _, err := websocket.DefaultDialer.Dial("wss://gateway.discord.gg/?v=10&encoding=json", nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()

	users := []string{}
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return nil
		}

		var response map[string]interface{}
		if err := json.Unmarshal(message, &response); err != nil {
			log.Println("unmarshal:", err)
			return nil
		}

		if response["t"] == nil {
			sendData := map[string]interface{}{
				"op": 2,
				"d": map[string]interface{}{
					"token":        token,
					"capabilities": 16381,
					"properties": map[string]interface{}{
						"os":                       "Android",
						"browser":                  "Discord Android",
						"device":                   "Android",
						"system_locale":            "ja-JP",
						"browser_user_agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
						"browser_version":          "122.0.0.0",
						"os_version":               "",
						"referrer":                 "",
						"referring_domain":         "",
						"referrer_current":         "",
						"referring_domain_current": "",
						"release_channel":          "stable",
						"client_build_number":      263582,
						"client_event_source":      nil,
					},
					"presence": map[string]interface{}{
						"status":     "invisible",
						"since":      0,
						"activities": []interface{}{},
						"afk":        false,
					},
					"compress": false,
					"client_state": map[string]interface{}{
						"guild_versions":              map[string]interface{}{},
						"highest_last_message_id":     "0",
						"read_state_version":          0,
						"user_guild_settings_version": -1,
						"private_channels_version":    "0",
						"api_code_version":            0,
					},
				},
			}
			if err := conn.WriteJSON(sendData); err != nil {
				log.Println("write:", err)
				return nil
			}
		} else if response["t"].(string) == "READY_SUPPLEMENTAL" {
			sendData := map[string]interface{}{
				"op": 14,
				"d": map[string]interface{}{
					"guild_id":   server,
					"typing":     true,
					"activities": true,
					"threads":    true,
					"channels": map[string]interface{}{
						channel: [][]int{{0, 99}, {100, 199}, {200, 299}},
					},
				},
			}
			if err := conn.WriteJSON(sendData); err != nil {
				log.Println("write:", err)
				return nil
			}
		} else if response["t"].(string) == "GUILD_MEMBER_LIST_UPDATE" {
			items := response["d"].(map[string]interface{})["ops"].([]interface{})
			for _, item := range items {
				for _, member := range item.(map[string]interface{})["items"].([]interface{}) {
					if member.(map[string]interface{})["member"] != nil {
						users = append(users, member.(map[string]interface{})["member"].(map[string]interface{})["user"].(map[string]interface{})["id"].(string))
					}
				}
			}
			return users
		}
	}
}
