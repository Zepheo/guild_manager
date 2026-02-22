package turtlogs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/zepheo/guild_manager/internal/domain"
)

type TurtlogsClient struct {
	BaseURL       string
	HTTPClient    *http.Client
	reID          *regexp.Regexp
	reTiny        *regexp.Regexp
	reFull        *regexp.Regexp
	reUnknownName *regexp.Regexp
}

func NewTurtlogsClient(baseURL string) *TurtlogsClient {
	escapedBase := regexp.QuoteMeta(baseURL)
	return &TurtlogsClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		reID:          regexp.MustCompile(`^[0-9]{5}$`),
		reTiny:        regexp.MustCompile(fmt.Sprintf(`^%s/tiny_url/([0-9]+)$`, escapedBase)),
		reFull:        regexp.MustCompile(fmt.Sprintf(`^%s/viewer/([0-9]+)/base.?$`, escapedBase)),
		reUnknownName: regexp.MustCompile("Unknown-[0-9]+"),
	}
}

type RaidData struct {
	Meta      *RaidMetaData
	Players   map[int]domain.Player
	LootDrops []domain.LootDrop
}

func (c *TurtlogsClient) GetRaidData(log string) (*RaidData, error) {
	id, err := c.getLogId(log)
	if err != nil {
		return nil, err
	}

	players, err := c.GetRaidPlayers(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get players: %v", err)
	}

	loot, err := c.GetRaidLoot(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get loot: %v", err)
	}

	meta, err := c.GetRaidMetaData(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get meta data: %v", err)
	}

	lootDrops := make([]domain.LootDrop, 0, len(loot))
	for _, drop := range loot {
		if winnerName, ok := players[drop.WinnerID]; ok {
			lootDrops = append(lootDrops, domain.LootDrop{
				ItemName:   drop.ItemName,
				WinnerName: winnerName.CharacterName,
			})
		} else {
			lootDrops = append(lootDrops, domain.LootDrop{
				ItemName:   drop.ItemName,
				WinnerName: "unknown",
			})
		}
	}

	return &RaidData{
		Meta:      meta,
		Players:   players,
		LootDrops: lootDrops,
	}, nil
}

type MillisecondTime struct {
	time.Time
}

func (m *MillisecondTime) UnmarshalJSON(b []byte) error {
	ms, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return err
	}

	m.Time = time.Unix(0, ms*int64(time.Millisecond))
	return nil
}

type RaidMetaData struct {
	RaidID   int             `json:"instance_meta_id"`
	RaidDate MillisecondTime `json:"start_ts"` // Use the custom type here
}

func (c *TurtlogsClient) GetRaidMetaData(log string) (*RaidMetaData, error) {
	endpoint := fmt.Sprintf("%s/API/instance/export/%s", c.BaseURL, log)
	resp, err := c.HTTPClient.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status: %d", resp.StatusCode)
	}

	var raidMetaData RaidMetaData
	if err := json.NewDecoder(resp.Body).Decode(&raidMetaData); err != nil {
		return nil, err
	}

	return &raidMetaData, nil
}

type Participants struct {
	CharacterID int    `json:"character_id"`
	Name        string `json:"name"`
	HeroClassID int    `json:"hero_class_id"`
}

func (c *TurtlogsClient) GetRaidPlayers(log string) (map[int]domain.Player, error) {
	endpoint := fmt.Sprintf("%s/API/instance/export/participants/%s", c.BaseURL, log)
	resp, err := c.HTTPClient.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status: %d", resp.StatusCode)
	}

	var participants []Participants
	if err := json.NewDecoder(resp.Body).Decode(&participants); err != nil {
		return nil, err
	}

	players := map[int]domain.Player{}

	for _, p := range participants {
		if p.HeroClassID < 12 && !c.reUnknownName.MatchString(p.Name) {
			players[p.CharacterID] = domain.Player{
				ID:            p.CharacterID,
				CharacterName: p.Name,
				Class:         mapClassId(p.HeroClassID),
			}
		}
	}

	return players, nil
}

type Loot struct {
	ItemName string
	WinnerID int
}

func (c *TurtlogsClient) GetRaidLoot(log string) ([]Loot, error) {
	endpoint := fmt.Sprintf("%s/API/instance/export/%s/3/0", c.BaseURL, log)
	resp, err := c.HTTPClient.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status: %d", resp.StatusCode)
	}

	var rawLoot [][]any
	if err := json.NewDecoder(resp.Body).Decode(&rawLoot); err != nil {
		return nil, err
	}

	lootDrops := make([]Loot, 0, len(rawLoot))

	for _, entry := range rawLoot {
		if len(entry) < 8 {
			continue
		}
		itemName := entry[5].(string)
		winnerID := int(entry[3].(float64))
		lootDrops = append(lootDrops, Loot{
			ItemName: itemName,
			WinnerID: winnerID,
		})
	}
	return lootDrops, nil
}

func (c *TurtlogsClient) getLogId(log string) (string, error) {
	logType, err := c.identifyLog(log)
	if err != nil {
		return "", err
	}

	var matches []string
	switch *logType {
	case LogTypeId:
		return log, nil
	case LogTypeTiny:
		matches = c.reTiny.FindStringSubmatch(log)
		if len(matches) > 1 {
			return c.resolveTinyID(matches[1])
		}
	case LogTypeFull:
		matches := c.reFull.FindStringSubmatch(log)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}
	return "", fmt.Errorf("Could not extract ID from log")
}

type LogType int

const (
	LogTypeId LogType = iota
	LogTypeTiny
	LogTypeFull
)

func (c *TurtlogsClient) identifyLog(log string) (*LogType, error) {
	var logType LogType
	switch {
	case c.reID.MatchString(log):
		logType = LogTypeId
		return &logType, nil
	case c.reTiny.MatchString(log):
		logType = LogTypeTiny
		return &logType, nil
	case c.reFull.MatchString(log):
		logType = LogTypeFull
		return &logType, nil
	default:
		return nil, fmt.Errorf("Log does not match a known format")
	}
}

type TinyResolverResponse struct {
	ID         int    `json:"id"`
	URLPayload string `json:"url_payload"` // This is a string of JSON
}

type InnerPayload struct {
	Payload struct {
		InstanceMetaID int `json:"instance_meta_id"` // This is your 90027
	} `json:"payload"`
}

func (c *TurtlogsClient) resolveTinyID(tinyId string) (string, error) {
	endpoint := fmt.Sprintf("%s/API/utility/tiny_url/%s", c.BaseURL, tinyId)
	resp, err := c.HTTPClient.Get(endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result TinyResolverResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("error in decode")
		return "", err
	}

	var inner InnerPayload
	if err := json.Unmarshal([]byte(result.URLPayload), &inner); err != nil {
		fmt.Println("error in unmarshal")
		return "", err
	}

	return fmt.Sprint(inner.Payload.InstanceMetaID), nil
}
