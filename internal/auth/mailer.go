package auth

import (
	"fmt"
	"log"
	"net/smtp"
)

// Mailer 邮件发送接口
type Mailer interface {
	SendVerification(toEmail, nickname, verifyURL string) error
}

// SMTPConfig SMTP 配置（与 internal/config.SMTPConfig 同构，避免循环引用）
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	DevMode  bool
}

// smtpMailer 真实 SMTP 发送
type smtpMailer struct {
	cfg SMTPConfig
}

// devMailer 开发模式：打印到日志
type devMailer struct{}

// NewMailer 根据配置创建 Mailer
func NewMailer(cfg SMTPConfig) Mailer {
	if cfg.DevMode || cfg.Host == "" {
		log.Println("[mailer] dev mode enabled: verification emails will be printed to log")
		return &devMailer{}
	}
	if cfg.From == "" {
		cfg.From = cfg.Username
	}
	return &smtpMailer{cfg: cfg}
}

// SendVerification 实现
func (m *smtpMailer) SendVerification(toEmail, nickname, verifyURL string) error {
	subject := "【花好月圆 f3moon】请验证您的邮箱"
	body := fmt.Sprintf(`%s 你好，

欢迎注册 f3moon (花好月圆)。

请点击以下链接完成邮箱验证（24 小时内有效）：

%s

如果不是您本人操作，请忽略此邮件。

— f3moon
`, displayName(nickname, toEmail), verifyURL)

	msg := buildMessage(m.cfg.From, toEmail, subject, body)
	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)
	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	if err := smtp.SendMail(addr, auth, m.cfg.From, []string{toEmail}, msg); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}
	return nil
}

func (m *devMailer) SendVerification(toEmail, nickname, verifyURL string) error {
	log.Printf("[mailer:dev] => %s (%s)\n  验证链接: %s", toEmail, nickname, verifyURL)
	return nil
}

func displayName(nickname, email string) string {
	if nickname != "" {
		return nickname
	}
	return email
}

func buildMessage(from, to, subject, body string) []byte {
	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n",
		from, to, subject)
	return []byte(headers + body)
}
