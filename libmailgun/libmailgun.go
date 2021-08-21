package libmailgun

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/helloferdie/stdgo/logger"
	"github.com/mailgun/mailgun-go/v4"
)

// Recipient -
type Recipient struct {
	Email     string
	Variables map[string]interface{}
}

// SendSimple -
func SendSimple(sender string, recipient []Recipient, subject string, content string, attachment []string) error {
	mg := mailgun.NewMailgun(os.Getenv("mailgun_domain"), os.Getenv("mailgun_api_key"))

	if sender == "" {
		sender = os.Getenv("mailgun_sender")
	}

	msg := mg.NewMessage(sender, subject, content)
	if len(recipient) <= 0 {
		return fmt.Errorf("%s", "mailgun.error_empty_recipient_list")
	}

	for _, rc := range recipient {
		msg.AddRecipientAndVariables(rc.Email, rc.Variables)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, id, err := mg.Send(ctx, msg)
	if err != nil {
		logger.MakeLogEntry(nil, true).Errorf("Error mailgun - send simple (%v) error: %v", id, err)
		return fmt.Errorf("%s", "mailgun.error_send_simple")
	}
	return nil
}

// SendTemplate -
func SendTemplate(sender string, recipient *Recipient, bcc []string, subject string, template string, attachment []string) error {
	mg := mailgun.NewMailgun(os.Getenv("mailgun_domain"), os.Getenv("mailgun_api_key"))

	if sender == "" {
		sender = os.Getenv("mailgun_sender")
	}

	msg := mg.NewMessage(sender, subject, "")
	msg.SetTemplate(template)
	for _, v := range bcc {
		msg.AddBCC(v)
	}
	msg.AddRecipient(recipient.Email)

	vars, err := json.Marshal(recipient.Variables)
	if err != nil {
		logger.MakeLogEntry(nil, true).Errorf("Error mailgun - mashal template variables error: %v", err)
		return fmt.Errorf("%s", "mailgun.error_marshal_template_variables")
	}
	msg.AddHeader("X-Mailgun-Variables", string(vars))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, id, err := mg.Send(ctx, msg)
	if err != nil {
		logger.MakeLogEntry(nil, true).Errorf("Error mailgun - send template (%v) error: %v", id, err)
		return fmt.Errorf("%s", "mailgun.error_send_template")
	}
	return nil
}
