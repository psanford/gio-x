//+build android

package notify

import (
	"gioui.org/x/notify/android"
)

type androidManager struct {
	channel *android.NotificationChannel
}

var _ Manager = &androidManager{}

func newManager() (Manager, error) {
	channel, err := android.NewChannel(android.ImportanceDefault, "DEFAULT", "niotify", "background notifications")
	if err != nil {
		return Manager{}, err
	}
	return &androidManager{
		channel: channel,
	}, nil
}

func (a *androidManager) CreateNotification(title, text string) (Notification, error) {
	notification, err := a.channel.Send(title, text)
	if err != nil {
		return nil, err
	}
	return &notification, nil
}
