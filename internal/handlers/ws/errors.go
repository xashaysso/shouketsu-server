package ws

import (
	"encoding/json"
	"errors"
	"shouketsu/internal/domain"
)

var (
	ErrNonExistentRoom = errors.New("This room doesn't exist")
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
