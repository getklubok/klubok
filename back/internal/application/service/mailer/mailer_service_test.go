package mailer

import (
	"errors"
	"html/template"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"klubok/internal/application/service/mailer/mocks"
)

func TestMailer_Send(t *testing.T) {
	ctx := t.Context()

	type args struct {
		recipients []string
		subject    string
		tpl        Template
		content    map[string]any
	}
	tests := []struct {
		name    string
		service func() *Mailer
		args    args
		err     error
	}{
		{
			name: "send mail",
			service: func() *Mailer {
				client := mocks.NewMailClientMock(t)

				client.EXPECT().
					DialAndSendWithContext(mock.Anything, mock.Anything).
					Return(nil)

				return &Mailer{
					from:   "MyApp <no-reply@domain.tld>",
					client: client,
				}
			},
			args: args{
				recipients: []string{"test@domain.tld"},
				subject:    "Auth code",
				tpl:        TemplateAuthCode,
				content: map[string]any{
					"code":       "123456",
					"timeString": "5 минут",
				},
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.service().Send(ctx, tt.args.recipients, tt.args.subject, tt.args.tpl, tt.args.content)
			assert.Equal(t, tt.err, err)
		})
	}
}

func Test_getTemplate(t *testing.T) {
	type args struct {
		tpl Template
	}
	tests := []struct {
		name string
		args args
		want *template.Template
		err  error
	}{
		{
			name: "auth code template",
			args: args{
				tpl: TemplateAuthCode,
			},
			want: template.Must(template.New("html_template").ParseFS(efs, LayoutBase.String(), TemplateAuthCode.String())),
			err:  nil,
		},
		{
			name: "successful registration template",
			args: args{
				tpl: TemplateRegisterSuccess,
			},
			want: template.Must(template.New("html_template").ParseFS(efs, LayoutBase.String(), TemplateRegisterSuccess.String())),
			err:  nil,
		},
		{
			name: "unknown template",
			args: args{
				tpl: "unknown",
			},
			want: nil,
			err:  errors.New("unknown template: unknown"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getTemplate(tt.args.tpl)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.err, err)
		})
	}
}

func Test_assembleMessage(t *testing.T) {
	type args struct {
		from       string
		recipients []string
		subject    string
	}
	type result struct {
		from  []string
		to    []string
		fname string
	}
	tests := []struct {
		name        string
		htpl        Template
		data        map[string]any
		args        args
		fnameActual string
		want        result
	}{
		{
			name: "assemble message: auth code template",
			htpl: TemplateAuthCode,
			data: map[string]any{
				"code":       "123456",
				"timeString": "5 минут",
			},
			args: args{
				from:       "MyApp <no-reply@domain.tld>",
				recipients: []string{"test@domain.tld"},
				subject:    "Auth code",
			},
			fnameActual: "testdata/auth_code.actual.html",
			want: result{
				from:  []string{`"MyApp" <no-reply@domain.tld>`},
				to:    []string{"<test@domain.tld>"},
				fname: "testdata/auth_code.expected.html",
			},
		},
		{
			name: "assemble message: successful registration template",
			htpl: TemplateRegisterSuccess,
			data: map[string]any{
				"sysAppURL":    "https://domain.tld",
				"userID":       "01J8SCNHRR83ANJVPV3BKKE61H",
				"activateCode": "C4rKm3sT5vWxYz2A4D6F7H9KMNPRSTUVWXYZacde",
				"timeString":   "7 дней",
			},
			args: args{
				from:       "MyApp <no-reply@domain.tld>",
				recipients: []string{"test@domain.tld"},
				subject:    "Successful registration",
			},
			fnameActual: "testdata/register_success.actual.html",
			want: result{
				from:  []string{`"MyApp" <no-reply@domain.tld>`},
				to:    []string{"<test@domain.tld>"},
				fname: "testdata/register_success.expected.html",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			htpl, err := getTemplate(tt.htpl)
			assert.NoError(t, err)

			content, err := assembleContent(htpl, tt.data)
			assert.NoError(t, err)

			got, err := assembleMessage(tt.args.from, tt.args.recipients, tt.args.subject, content)
			assert.NoError(t, err)
			assert.Equal(t, tt.want.from, got.GetFromString())
			assert.Equal(t, tt.want.to, got.GetToString())

			actual, err := got.GetParts()[0].GetContent()
			assert.NoError(t, err)
			contentToFile(t, tt.fnameActual, actual)

			expected := contentFromFile(t, tt.want.fname)
			assert.Equal(t, string(expected), string(actual))
		})
	}
}

func contentToFile(t *testing.T, fpath string, content []byte) {
	t.Helper()

	err := os.WriteFile(fpath, content, 0600)
	require.NoError(t, err)
}

func contentFromFile(t *testing.T, fpath string) []byte {
	t.Helper()

	b, err := os.ReadFile(fpath)
	require.NoError(t, err)

	return b
}
