package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var users []map[string]interface{}
var usersMutex sync.Mutex

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
		log.Printf("[VALID]   %v [%v]\n", token, response)
	} else {
		log.Printf("[INVALID] %v [%v]", token, response)
		os.Exit(1)
	}
	fmt.Printf("Scraping in %s with %s\n", serverID, token)
	members := scrape(token, serverID, channelID)
	fmt.Printf("Total Scrapped: %d\n", len(members))
	userIDs := make([]string, len(members))
	for i, member := range members {
		userIDs[i] = member["id"].(string)
	}

	fmt.Printf("Scrapped Id: %v\n", userIDs)
}

func checktoken(token string) int {
	url := "https://discord.com/api/v9/users/@me/affinities/guilds"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("%v", err)
		return -1
	}
	req.Header.Set("authorization", token)
	httpClient := &http.Client{}
	resp, err := httpClient.Do(commonheaders(req))
	if err != nil {
		log.Printf("%v", err)
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

func rangeCorrector(ranges [][]int) [][]int {
	if !containsRange(ranges, []int{0, 99}) {
		ranges = append([][]int{{0, 99}}, ranges...)
	}
	return ranges
}

func getRanges(index, multiplier, memberCount int) [][]int {
	initialNum := index * multiplier
	rangesList := [][]int{{initialNum, initialNum + 99}}
	if memberCount > initialNum+99 {
		rangesList = append(rangesList, []int{initialNum + 100, initialNum + 199})
	}
	return rangeCorrector(rangesList)
}

func parseGuildMemberListUpdate(response map[string]interface{}) map[string]interface{} {
	memberdata := map[string]interface{}{
		"online_count":  response["d"].(map[string]interface{})["online_count"],
		"member_count":  response["d"].(map[string]interface{})["member_count"],
		"id":            response["d"].(map[string]interface{})["id"],
		"guild_id":      response["d"].(map[string]interface{})["guild_id"],
		"hoisted_roles": response["d"].(map[string]interface{})["groups"],
		"types":         []interface{}{},
		"locations":     []interface{}{},
		"updates":       []interface{}{},
	}

	for _, chunk := range response["d"].(map[string]interface{})["ops"].([]interface{}) {
		memberdata["types"] = append(memberdata["types"].([]interface{}), chunk.(map[string]interface{})["op"])
		if chunk.(map[string]interface{})["op"] == "SYNC" || chunk.(map[string]interface{})["op"] == "INVALIDATE" {
			memberdata["locations"] = append(memberdata["locations"].([]interface{}), chunk.(map[string]interface{})["range"])
			if chunk.(map[string]interface{})["op"] == "SYNC" {
				memberdata["updates"] = append(memberdata["updates"].([]interface{}), chunk.(map[string]interface{})["items"])
			} else {
				memberdata["updates"] = append(memberdata["updates"].([]interface{}), []interface{}{})
			}
		} else if chunk.(map[string]interface{})["op"] == "INSERT" || chunk.(map[string]interface{})["op"] == "UPDATE" || chunk.(map[string]interface{})["op"] == "DELETE" {
			memberdata["locations"] = append(memberdata["locations"].([]interface{}), chunk.(map[string]interface{})["index"])
			if chunk.(map[string]interface{})["op"] == "DELETE" {
				memberdata["updates"] = append(memberdata["updates"].([]interface{}), []interface{}{})
			} else {
				memberdata["updates"] = append(memberdata["updates"].([]interface{}), chunk.(map[string]interface{})["item"])
			}
		}
	}

	return memberdata
}

type DiscordSocket struct {
	Token         string
	GuildID       string
	ChannelID     string
	SocketHeaders map[string]string
	Conn          *websocket.Conn
	EndScraping   bool
	Guilds        map[string]map[string]interface{}
	Members       map[string]interface{}
	Ranges        [][]int
	LastRange     int
	PacketsRecv   int
	Mutex         sync.Mutex
}

func (ds *DiscordSocket) run() {
	ds.runForever()
}

func (ds *DiscordSocket) scrapeUsers() {
	log.Println("scrapeUsers called")
	if !ds.EndScraping {
		log.Println("Sending scrape request...")
		rangesJSON, err := json.Marshal(ds.Ranges)
		if err != nil {
			log.Println("Error marshalling Ranges:", err)
			return
		}
		fmt.Println(rangesJSON)

		requestJSON1 := `{"op":14,"d":{"guild_id":"` + ds.GuildID + `","typing":true,"activities":true,"threads":true,"channels":{"` + ds.ChannelID + `":` + string(rangesJSON) + `}}}`
		fmt.Println(requestJSON1)
		ds.send(requestJSON1)
		requestJSON2 := `{"op":14,"d":{"guild_id":"` + ds.GuildID + `","typing":true,"activities":true,"threads":true,"channels":{"` + ds.ChannelID + `":[[0, 99], [100, 199]]}}}`
		fmt.Println(requestJSON2)
		ds.send(requestJSON2)
	}
}

func (ds *DiscordSocket) sockOpen() {
	ds.send(fmt.Sprintf(`{"op":2,"d":{"token":"%s","capabilities":125,"properties":{"os":"Windows","browser":"Firefox","device":"","system_locale":"it-IT","browser_user_agent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:94.0) Gecko/20100101 Firefox/94.0","browser_version":"94.0","os_version":"10","referrer":"","referring_domain":"","referrer_current":"","referring_domain_current":"","release_channel":"stable","client_build_number":103981,"client_event_source":null},"presence":{"status":"online","since":0,"activities":[],"afk":false},"compress":false,"client_state":{"guild_hashes":{},"highest_last_message_id":"0","read_state_version":0,"user_guild_settings_version":-1,"user_settings_version":-1}}}`, ds.Token))
}

func (ds *DiscordSocket) heartbeatThread(interval time.Duration) {
	for {
		ds.send(fmt.Sprintf(`{"op":1,"d":%d}`, ds.PacketsRecv))
		time.Sleep(interval)
	}
}

func (ds *DiscordSocket) sockMessage(message []byte) {
	//log.Println("Received message:", string(message))

	var decoded map[string]interface{}
	err := json.Unmarshal(message, &decoded)
	if err != nil || decoded == nil {
		log.Println("Error decoding message:", err)
		return
	}

	ds.Mutex.Lock()
	defer ds.Mutex.Unlock()

	if decoded["op"].(float64) != 11 {
		ds.PacketsRecv++
	}

	switch decoded["op"].(float64) {
	case 10:
		log.Println("Received OP 10 (Hello)")
		go ds.heartbeatThread(time.Duration(decoded["d"].(map[string]interface{})["heartbeat_interval"].(float64)) * time.Millisecond)
	case 0:
		switch decoded["t"].(string) {
		case "READY":
			log.Println("Received READY event")
			for _, guild := range decoded["d"].(map[string]interface{})["guilds"].([]interface{}) {
				guildID := guild.(map[string]interface{})["id"].(string)
				ds.Guilds[guildID] = map[string]interface{}{
					"member_count": guild.(map[string]interface{})["member_count"].(float64),
				}
				log.Println("Added Guild ID to ds.Guilds:", guildID)
				guild, ok := ds.Guilds[guildID]
				if !ok || guild == nil {
					log.Printf("Error: Guild not found for ID %s or it is nil\n", guildID)
					os.Exit(1)
					//return
				}
			}
		case "READY_SUPPLEMENTAL":
			if decoded["t"].(string) == "READY_SUPPLEMENTAL" {
				log.Println("Received READY_SUPPLEMENTAL event") // この行を追加
				//ds.Mutex.Lock()
				//defer ds.Mutex.Unlock()
				guildID := ds.GuildID
				if guildID == "" {
					log.Println("Error: GuildID is empty")
					return
				}

				guild, ok := ds.Guilds[guildID]
				if !ok || guild == nil {
					log.Printf("Error: Guild not found for ID %s or it is nil\n", guildID)
					//os.Exit(1)
					//return
				}

				memberCount, ok := guild["member_count"].(float64)
				if !ok {
					log.Println("Error: Unable to get member count from Guild")
					//return
				}

				ds.Ranges = getRanges(0, 100, int(memberCount))
				ds.scrapeUsers()
			}
		case "GUILD_MEMBER_LIST_UPDATE":
			log.Println("Received GUILD_MEMBER_LIST_UPDATE event")
			parsed := parseGuildMemberListUpdate(decoded)
			if parsed["guild_id"].(string) == ds.GuildID && (containsType(parsed["types"].([]interface{}), "SYNC") || containsType(parsed["types"].([]interface{}), "UPDATE")) {
				for elem, index := range parsed["types"].([]interface{}) {
					if index == "SYNC" {
						log.Println("Processing SYNC")
						if len(parsed["updates"].([]interface{})[elem].([]interface{})) == 0 {
							ds.EndScraping = true
							log.Println("No updates in SYNC, ending scraping.")
							break
						}
						for _, item := range parsed["updates"].([]interface{})[elem].([]interface{}) {
							if item.(map[string]interface{})["member"] != nil {
								mem := item.(map[string]interface{})["member"].(map[string]interface{})
								obj := map[string]interface{}{
									"tag": mem["user"].(map[string]interface{})["username"].(string) + "#" + mem["user"].(map[string]interface{})["discriminator"].(string),
									"id":  mem["user"].(map[string]interface{})["id"].(string),
								}
								usersMutex.Lock()
								users = append(users, obj)
								usersMutex.Unlock()
							}
						}
					} else if index == "UPDATE" {
						log.Println("Processing UPDATE")
						for _, item := range parsed["updates"].([]interface{})[elem].([]interface{}) {
							if item.(map[string]interface{})["member"] != nil {
								mem := item.(map[string]interface{})["member"].(map[string]interface{})
								obj := map[string]interface{}{
									"tag": mem["user"].(map[string]interface{})["username"].(string) + "#" + mem["user"].(map[string]interface{})["discriminator"].(string),
									"id":  mem["user"].(map[string]interface{})["id"].(string),
								}
								usersMutex.Lock()
								users = append(users, obj)
								usersMutex.Unlock()
							}
						}
					}
					ds.LastRange++
					ds.Ranges = getRanges(ds.LastRange, 100, int(ds.Guilds[ds.GuildID]["member_count"].(float64)))
					time.Sleep(350 * time.Millisecond)
					ds.scrapeUsers()
				}
			}
			if ds.EndScraping {
				log.Println("End scraping, closing connection.")
				ds.Conn.Close()
			}
		}
	}
}

func scrape(token, guildID, channelID string) []map[string]interface{} {
	ds := &DiscordSocket{
		Token:         token,
		GuildID:       guildID,
		ChannelID:     channelID,
		SocketHeaders: map[string]string{"Accept-Encoding": "gzip, deflate, br", "Accept-Language": "en-US,en;q=0.9", "Cache-Control": "no-cache", "Pragma": "no-cache", "Sec-WebSocket-Extensions": "permessage-deflate; client_max_window_bits", "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.1.2 Safari/605.1.15"},
		Guilds:        make(map[string]map[string]interface{}),
		Members:       make(map[string]interface{}),
		Ranges:        [][]int{{0, 0}},
		LastRange:     0,
		PacketsRecv:   0,
		Mutex:         sync.Mutex{},
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		ds.run()
	}()

	<-done
	return users
}

func containsRange(ranges [][]int, target []int) bool {
	for _, r := range ranges {
		if r[0] == target[0] && r[1] == target[1] {
			return true
		}
	}
	return false
}

func containsType(types []interface{}, target string) bool {
	for _, t := range types {
		if t.(string) == target {
			return true
		}
	}
	return false
}

func (ds *DiscordSocket) runForever() {
	url := "wss://gateway.discord.gg/?encoding=json&v=9"
	conn, _, err := websocket.DefaultDialer.Dial(url, http.Header{"Sec-WebSocket-Extensions": {"permessage-deflate; client_max_window_bits"}})
	if err != nil {
		log.Fatal("Error connecting to WebSocket:", err)
	}
	ds.Conn = conn

	log.Println("WebSocket connection established.")

	go ds.sockOpen()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error reading message:", err)
				return
			}
			go ds.sockMessage(message)
		}
	}()

	<-done
}

func (ds *DiscordSocket) send(message string) {
	//ds.Mutex.Lock()
	//defer ds.Mutex.Unlock()

	err := ds.Conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Println("Error writing message:", err)
		return
	}
	log.Println("Sent message:", message)
}
