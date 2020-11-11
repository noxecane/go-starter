package notification

type MailerMock struct {
}

func (n *MailerMock) Send(m TemplateMail) error {
	return nil
}
