
package users

import "time"

type User struct {
	UUID      string    `json:"uuid"`
	CreatedAt time.Time `json:"created_at"`
	// Future: per-user egress, rate limits, notes, tags
}
