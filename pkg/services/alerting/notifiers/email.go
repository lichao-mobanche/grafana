package notifiers

import (
	"strings"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/log"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/alerting"
)

func init() {
	alerting.RegisterNotifier("email", NewEmailNotifier)
}

type EmailNotifier struct {
	NotifierBase
	Addresses []string
	log       log.Logger
}

func NewEmailNotifier(model *m.AlertNotification) (alerting.Notifier, error) {
	addressesString := model.Settings.Get("addresses").MustString()

	if addressesString == "" {
		return nil, alerting.AlertValidationError{Reason: "Could not find addresses in settings"}
	}

	return &EmailNotifier{
		NotifierBase: NotifierBase{
			Name: model.Name,
			Type: model.Type,
		},
		Addresses: strings.Split(addressesString, "\n"),
		log:       log.New("alerting.notifier.email"),
	}, nil
}

func (this *EmailNotifier) Notify(context *alerting.AlertResultContext) {
	this.log.Info("Sending alert notification to", "addresses", this.Addresses)

	ruleLink, err := getRuleLink(context.Rule)
	if err != nil {
		this.log.Error("Failed get rule link", "error", err)
		return
	}

	cmd := &m.SendEmailCommand{
		Data: map[string]interface{}{
			"RuleState": context.Rule.State,
			"RuleName":  context.Rule.Name,
			"Severity":  context.Rule.Severity,
			"RuleLink":  ruleLink,
		},
		To:       this.Addresses,
		Template: "alert_notification.html",
	}

	if err := bus.Dispatch(cmd); err != nil {
		this.log.Error("Failed to send alert notification email", "error", err)
	}
}
