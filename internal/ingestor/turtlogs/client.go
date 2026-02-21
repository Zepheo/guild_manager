package turtlogs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type TurtlogsClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewTurtlogsClient(baseURL string) *TurtlogsClient {
	return &TurtlogsClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type LootEntry struct {
	ItemName   string `json:"item_name"`
	PlayerName string `json:"player_name"`
}

type LogResponse struct {
	Loot []LootEntry `json:"loot"`
}

func (c *TurtlogsClient) GetRaidLoot(logID string) (*LogResponse, error) {
	url := fmt.Sprintf("%s/api/v1/logs/%s", c.BaseURL, logID)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call turtlogs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("turtlogs returned error status: %d", resp.StatusCode)
	}

	var logData LogResponse
	if err := json.NewDecoder(resp.Body).Decode(&logData); err != nil {
		return nil, fmt.Errorf("failed to decode turtlogs response: %w", err)
	}

	return &logData, nil
}
