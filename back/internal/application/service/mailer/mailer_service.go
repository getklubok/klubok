package mailer

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"maps"

	"github.com/wneessen/go-mail"
)

//go:embed logo.png templates/*
var efs embed.FS

// Mailer.
type Mailer struct {
	from    string
	client  MailClient
	sysData map[string]any
}

// New Mailer.
func New(from string, client MailClient, sysData map[string]any) *Mailer {
	return &Mailer{
		from:    from,
		client:  client,
		sysData: sysData,
	}
}

// Отправить письмо.
func (m *Mailer) Send(ctx context.Context, recipients []string, subject string, tpl Template, data map[string]any) error {
	htpl, err := getTemplate(tpl)
	if err != nil {
		return fmt.Errorf("get template: %w", err)
	}

	maps.Copy(data, m.sysData)
	content, err := assembleContent(htpl, data)
	if err != nil {
		return fmt.Errorf("assemble content: %w", err)
	}

	msg, err := assembleMessage(m.from, recipients, subject, content)
	if err != nil {
		return fmt.Errorf("assemble message: %w", err)
	}

	if err := m.client.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}

	return nil
}

// getTemplate возвращает HTML-шаблон.
func getTemplate(tpl Template) (*template.Template, error) {
	switch tpl {
	case TemplateAuthCode, TemplateRegisterSuccess:
		// ok

	default:
		return nil, fmt.Errorf("unknown template: %s", tpl)
	}

	htpl, err := template.New("html_template").ParseFS(efs, LayoutBase.String(), tpl.String())
	if err != nil {
		return nil, fmt.Errorf("failed to parse html template: %w", err)
	}

	return htpl, nil
}

// assembleContent собирает содержимое письма.
func assembleContent(htpl *template.Template, data map[string]any) (string, error) {
	var buf bytes.Buffer

	err := htpl.ExecuteTemplate(&buf, LayoutBaseFilename, data)
	if err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

func assembleMessage(
	from string,
	recipients []string,
	subject string,
	htmlContent string,
) (*mail.Msg, error) {
	msg := mail.NewMsg()

	if err := msg.From(from); err != nil {
		return nil, fmt.Errorf("failed to set From address: %w", err)
	}

	if err := msg.To(recipients...); err != nil {
		return nil, fmt.Errorf("failed to set To address: %w", err)
	}

	if err := msg.EmbedFromEmbedFS("logo.png", &efs, mail.WithFileContentID("logo.png"), mail.WithFileContentType("image/png")); err != nil {
		return nil, fmt.Errorf("failed to embed logo: %w", err)
	}

	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextHTML, htmlContent)

	return msg, nil
}
