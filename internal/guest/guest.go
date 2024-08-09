package guest

import (
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
)

type Guest struct {
	ID        uuid.UUID
	Message   string
	CreatedAt time.Time
	IP        net.IP
}

func NewGuest(message string, ip net.IP) (Guest, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Guest{}, fmt.Errorf("failed to create guest: %w", err)
	}

	return Guest{
		ID:        id,
		Message:   message,
		CreatedAt: time.Now(),
		IP:        ip,
	}, nil
}
