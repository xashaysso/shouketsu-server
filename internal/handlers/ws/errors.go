package ws

import (
	"encoding/json"
	"shouketsu/internal/domain"
)

func SendError(client *domain.Client, message string) {
	bytes, _ := json.Marshal(map[string]any{
		"type": "ERROR",
		"payload": map[string]string{
			"message": message,
		},
	})
	select {
	case client.Send <- bytes:
	default:
	}

	close(client.Send)
}
