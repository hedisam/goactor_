package pid

import (
	"github.com/hedisam/goactor/internal/mailbox"
)

type PID interface {
	Mailbox() mailbox.Mailbox
}
