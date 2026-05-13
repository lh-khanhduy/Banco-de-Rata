package mail

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSendEmail(t *testing.T) {
	mailer, err := NewMailerFromConfig()
	require.NoError(t, err)

	to := []string{"khanhduywin1907@gmail.com"}
	subject := "Hello from go-mail"
	plain := "Hi,\n\nThis is a test email sent via go-mail.\n\nRegards,\nSender"
	html := `<h1>Hello world</h1>
	<p>This is a test message from Willies</p>`
	attachFiles := []string{"../start.sh"}

	err = mailer.SendEmail(to, subject, plain, html, attachFiles)
	require.NoError(t, err)
}
