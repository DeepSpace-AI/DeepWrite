# mailer

通用邮件发送包，支持：
- SMTP 发送
- HTML 模板渲染（字符串模板 / 文件模板）
- `text/plain` 与 `text/html` 多段邮件

## 快速使用

```go
package main

import (
    "context"
    "github.com/deepwrite/serivces/gateway/pkg/mailer"
)

func send() error {
    sender, err := mailer.NewSMTPSender(mailer.SMTPConfig{
        Host:     "smtp.example.com",
        Port:     587,
        Username: "no-reply@example.com",
        Password: "your-password",
        FromName: "DeepWrite",
        FromMail: "no-reply@example.com",
    })
    if err != nil {
        return err
    }

    html, err := mailer.RenderHTMLTemplateString(`
        <h1>Hello, {{.Name}}</h1>
        <p>Your code is <b>{{.Code}}</b></p>
    `, map[string]any{
        "Name": "Landon",
        "Code": "123456",
    })
    if err != nil {
        return err
    }

    return sender.Send(context.Background(), mailer.Message{
        To:      []string{"user@example.com"},
        Subject: "Welcome",
        TextBody: "Hello, your code is 123456",
        HTMLBody: html,
    })
}
```

## 模板渲染

```go
html, err := mailer.RenderHTMLTemplateFile("templates/welcome.html", data)
```

使用基础模板（layout + content）：

```go
html, err := mailer.RenderHTMLTemplateWithLayout(
    "templates/mail/base.html",
    "templates/mail/welcome.html",
    map[string]any{
        "Title": "欢迎加入",
        "Brand": "DeepWrite",
        "Name":  "Landon",
    },
)
```

> 模板基于 `html/template`。
