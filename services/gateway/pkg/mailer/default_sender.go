package mailer

import "sync"

var (
	defaultSender Sender
	senderMu      sync.RWMutex
)

func SetDefault(sender Sender) {
	senderMu.Lock()
	defer senderMu.Unlock()
	defaultSender = sender
}

func GetDefault() Sender {
	senderMu.RLock()
	defer senderMu.RUnlock()
	return defaultSender
}
