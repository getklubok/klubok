package mailer

const (
	LayoutBaseFilename = "_base.html"
)

type Layout string

const (
	LayoutBase Layout = "templates/_base.html"
)

func (t Layout) String() string {
	return string(t)
}

type Template string

const (
	TemplateAuthCode        Template = "templates/auth_code.html"
	TemplateRegisterSuccess Template = "templates/register_success.html"
)

func (t Template) String() string {
	return string(t)
}
