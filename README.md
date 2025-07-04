# Go-based SMTP Relay MTA

This is a lightweight, secure Mail Transfer Agent (MTA) built in Go, designed to accept mail from [Acelle Mail](https://acellemail.com/) or any SMTP client, queue it to disk, and relay it to remote SMTP servers using DNS MX resolution.

---

## 🔧 Features

* SMTP listener on port 587 with STARTTLS support
* User authentication via `users.txt`
* Mail queue system (writes `.eml` files to disk)
* Multi-threaded delivery workers with MX lookup
* STARTTLS-enabled SMTP delivery to remote servers

---

## 🧱 Project Structure

```
.
├── main.go          # SMTP server that receives and queues messages
├── queue.go         # Worker script that delivers queued messages via SMTP
├── /var/mailqueue/  # Mail spool directory containing .eml files
├── /root/users.txt  # Auth credentials (username:password)
├── /root/cert.pem   # TLS certificate for STARTTLS
├── /root/key.pem    # TLS private key
```

---

## 🔄 Mail Flow Diagram

```
[Acelle Mail / SMTP Client] 
       ↓ (SMTP AUTH + STARTTLS)
      [main.go → queues .eml]
       ↓
   [/var/mailqueue/*.eml]
       ↓
     [queue.go → MX Lookup]
       ↓ (STARTTLS)
  [Remote SMTP Server (MX)]
       ↓
     [Recipient Inbox]
```

---

## 🛡️ Security

| Layer       | Mechanism                          |
| ----------- | ---------------------------------- |
| SMTP Server | STARTTLS on port 587               |
| AUTH        | Username/password via users.txt    |
| Outbound    | STARTTLS if remote server supports |
| Filesystem  | Emails stored with 0644 perms      |
| Logs        | Delivery and error logs            |

---

## ⚙️ Usage

### Run the SMTP Server:

```bash
sudo go run main.go
```

### Run the Delivery Processor:

```bash
go run queue.go
```

---

## 📝 Example: users.txt

```
admin:strongpassword
newsletter:mailpass2024
```

---

## 📥 Example: Received Email

```
From: mail@connect.example.com
To: user@domain.com
Subject: Hello
Date: Fri, 04 Jul 2025 17:02:03 +0000
MIME-Version: 1.0
Content-Type: multipart/alternative; boundary="..."
... MIME body ...
```

---

## 🔁 Running as Services (Linux systemd)

### 🏗️ Build the binaries:

```bash
go build -o /usr/local/bin/mta-server main.go
go build -o /usr/local/bin/mta-queue queue.go
```

### 📄 Create systemd service units:

**/etc/systemd/system/mta-server.service**

```ini
[Unit]
Description=Go SMTP Server (MTA)
After=network.target

[Service]
ExecStart=/usr/local/bin/mta-server
WorkingDirectory=/root/go-mta
User=root
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

**/etc/systemd/system/mta-queue.service**

```ini
[Unit]
Description=Go SMTP Delivery Queue Processor
After=network.target

[Service]
ExecStart=/usr/local/bin/mta-queue
WorkingDirectory=/root/go-mta
User=root
Restart=always

[Install]
WantedBy=multi-user.target
```

### 🚀 Enable and start services:

```bash
sudo systemctl daemon-reexec
sudo systemctl daemon-reload

sudo systemctl enable mta-server.service
sudo systemctl enable mta-queue.service

sudo systemctl start mta-server.service
sudo systemctl start mta-queue.service
```

### ✅ Verify status:

```bash
sudo systemctl status mta-server.service
sudo systemctl status mta-queue.service
```

---

## 🚧 Roadmap

* [ ] Add retry logic for failed delivery
* [ ] DKIM header signing
* [ ] Web UI for queue monitoring
* [ ] Logging improvements with rotation

---

## 📜 License

MIT

---

## 🤝 Acknowledgements

* [Go-SMTP by emersion](https://github.com/emersion/go-smtp)
* [Acelle Mail](https://acellemail.com/)

---

> Built with ❤️ to deliver your campaigns reliably and securely.
