package invitations

import (
	"fmt"

	"tsaron.com/godview-starter/pkg/notification"
)

func SendInvitation(mailer notification.Mailer, route string, iv Invitation) error {
	data := struct {
		Route       string
		Token       string
		CompanyName string
	}{
		route,
		iv.Token,
		iv.CompanyName,
	}

	return mailer.Send(notification.TemplateMail{
		Sender:        notification.SenderPostmaster,
		Subject:       fmt.Sprintf("Invitation to %s", iv.CompanyName),
		ReceiverName:  "",
		ReceiverEmail: iv.EmailAddress,
		Template:      "invitation",
		TemplateData:  data,
	})
}
