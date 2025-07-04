// File: main.go
package main

import (
    "bufio"
    "crypto/tls"
    "fmt"
    "io"
    "log"
    "os"
    "strings"
    "time"

    smtp "github.com/emersion/go-smtp"
)

var (
    mailDir   = "/var/mailqueue"
    usersFile = "/root/users.txt"
    tlsCert   = "/root/cert.pem"
    tlsKey    = "/root/key.pem"
)

type Backend struct{}

func (bkd *Backend) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
    if !isValidUser(username, password) {
        log.Printf("❌ AUTH FAILED for %s\n", username)
        return nil, smtp.ErrAuthRequired
    }
    log.Printf("✅ AUTH SUCCESS for %s\n", username)
    return &Session{}, nil
}

func (bkd *Backend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
    return nil, smtp.ErrAuthRequired
}

type Session struct {
    from string
    to   string
}

func (s *Session) Mail(from string, opts smtp.MailOptions) error {
    s.from = strings.Trim(from, " <>")
    log.Println("MAIL FROM (cleaned):", s.from)
    return nil
}

func (s *Session) Rcpt(to string) error {
    s.to = strings.Trim(to, " <>")
    log.Println("RCPT TO (cleaned):", s.to)
    return nil
}


func (s *Session) Data(r io.Reader) error {
    body, err := io.ReadAll(r)
    if err != nil {
        return err
    }

    cleanFrom := strings.Trim(s.from, " <>")
    cleanTo := strings.Trim(s.to, " <>")




  full := string(body)



    filename := fmt.Sprintf("mail-%d.eml", time.Now().UnixNano())
    filepath := fmt.Sprintf("%s/%s", mailDir, filename)

    log.Printf("DEBUG: Saving mail with From: %s To: %s", cleanFrom, cleanTo)

    err = os.WriteFile(filepath, []byte(full), 0644)
    if err != nil {
        return err
    }

    log.Println("📩 Saved email to:", filepath)
    return nil
}




func (s *Session) Reset()         {}
func (s *Session) Logout() error { return nil }

func isValidUser(username, password string) bool {
    f, err := os.Open(usersFile)
    if err != nil {
        log.Println("❌ Error reading users file:", err)
        return false
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }

        parts := strings.SplitN(line, ":", 2)
        if len(parts) != 2 {
            continue
        }

        u, p := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
        if username == u && password == p {
            return true
        }
    }

    return false
}

func main() {
    backend := &Backend{}
    server := smtp.NewServer(backend)

    server.Addr = "0.0.0.0:587"
    server.Domain = "yourdomain.com"
    server.AuthDisabled = false
    server.AllowInsecureAuth = false

    cert, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
    if err != nil {
        log.Fatal("❌ Failed to load TLS cert/key:", err)
    }

    server.TLSConfig = &tls.Config{
        Certificates: []tls.Certificate{cert},
    }

    if _, err := os.Stat(mailDir); os.IsNotExist(err) {
        os.MkdirAll(mailDir, 0755)
    }

    log.Println("🚀 SMTP server running on", server.Addr)
    if err := server.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}
