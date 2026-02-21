package turtlogs

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type TurtlogsClient struct {
	BaseURL string
}

// LogResponse matches the JSON structure from Turtlogs
type LogResponse struct {
	Loot []struct {
		ItemName string `json:"item_name"`
		Player   string `json:"player_name"`
	} `json:"loot"`
}

func (c *TurtlogsClient) GetRaidLoot(logID string) (*LogResponse, error) {
	url := fmt.Sprintf("%s/api/v1/logs/%s", c.BaseURL, logID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var logData LogResponse
	if err := json.NewDecoder(resp.Body).Decode(&logData); err != nil {
		return nil, err
	}

	return &logData, nil
}
