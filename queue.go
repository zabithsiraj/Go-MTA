package main

import (
    "bufio"
    "fmt"
    "io/ioutil"
    "log"
    "net"
    "net/mail"
    "net/smtp"
    "os"
    "path/filepath"
    "strings"
    "sync"
)

func cleanEmailAddress(email string) string {
    return strings.Trim(email, " <>")
}



// 😼 Helo name check 
var heloHost string

func init() {
    out, err := exec.Command("hostname", "-f").Output()
    if err != nil {
        log.Fatalf("Failed to get FQDN: %v", err)
    }

    heloHost = strings.TrimSpace(string(out))
    log.Printf("Using HELO hostname: %s", heloHost)
}

const (
    queueDir    = "/var/mailqueue"
    numWorkers  = 10                      // 🔁 Adjust worker count here
)

func main() {
    files, err := filepath.Glob(filepath.Join(queueDir, "mail-*.eml"))
    if err != nil {
        log.Fatalf("Error reading queue directory: %v", err)
    }

    jobs := make(chan string, len(files))
    var wg sync.WaitGroup

    // Start workers
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for file := range jobs {
                if err := processMailFile(file); err != nil {
                    log.Printf("[Worker %d] Failed to process %s: %v", workerID, file, err)
                } else {
                    os.Remove(file)
                    log.Printf("[Worker %d] Delivered and removed: %s", workerID, file)
                }
            }
        }(i)
    }

    // Enqueue jobs
    for _, file := range files {
        jobs <- file
    }

    close(jobs)
    wg.Wait()
}

func processMailFile(filename string) error {
    content, err := ioutil.ReadFile(filename)
    if err != nil {
        return fmt.Errorf("read error: %w", err)
    }

    msg, err := mail.ReadMessage(strings.NewReader(string(content)))
    if err != nil {
        return fmt.Errorf("parse error: %w", err)
    }

   fromAddr, err := mail.ParseAddress(msg.Header.Get("From"))
if err != nil {
    return fmt.Errorf("invalid From address: %w", err)
}
toAddr, err := mail.ParseAddress(msg.Header.Get("To"))
if err != nil {
    return fmt.Errorf("invalid To address: %w", err)
}

from := fromAddr.Address
to := toAddr.Address





if from == "" || to == "" {
    return fmt.Errorf("missing From or To in headers")
}

recipients := strings.Split(to, ",")
for _, rcpt := range recipients {
    rcpt = cleanEmailAddress(rcpt)
    domain := strings.SplitN(rcpt, "@", 2)
    if len(domain) != 2 {
        log.Printf("Invalid recipient address: %s", rcpt)
        continue
    }

    mxRecords, err := net.LookupMX(domain[1])
    if err != nil || len(mxRecords) == 0 {
        return fmt.Errorf("MX lookup failed for %s: %w", domain[1], err)
    }
    

        mxHost := mxRecords[0].Host
        smtpAddr := fmt.Sprintf("%s:25", mxHost)

        log.Printf("Connecting to %s to deliver to %s", smtpAddr, rcpt)
        err = sendRawMail(from, []string{rcpt}, content, smtpAddr)
        if err != nil {
            return fmt.Errorf("delivery error: %w", err)
        }
    }

    return nil
}

func sendRawMail(from string, to []string, msg []byte, smtpAddr string) error {
    c, err := smtp.Dial(smtpAddr)
    if err != nil {
        return fmt.Errorf("dial failed: %w", err)
    }
    defer c.Close()

    if err := c.Hello(heloHost); err != nil {
        return fmt.Errorf("HELO failed: %w", err)
    }

    if err := c.Mail(from); err != nil {
        return fmt.Errorf("MAIL FROM failed: %w", err)
    }

    for _, addr := range to {
        if err := c.Rcpt(addr); err != nil {
            return fmt.Errorf("RCPT TO failed: %w", err)
        }
    }

    wc, err := c.Data()
    if err != nil {
        return fmt.Errorf("DATA failed: %w", err)
    }

    writer := bufio.NewWriter(wc)
    _, err = writer.Write(msg)
    if err != nil {
        return fmt.Errorf("write failed: %w", err)
    }
    writer.Flush()
    wc.Close()

    return c.Quit()
}
