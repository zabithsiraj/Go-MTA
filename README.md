# Go-based SMTP Relay MTA


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



---

> Ny Hobby work
