#!/bin/bash

set -e

# Constants
GO_VERSION="1.24.4"
GO_TAR="go$GO_VERSION.linux-amd64.tar.gz"
GO_URL="https://go.dev/dl/$GO_TAR"
GO_INSTALL_DIR="/usr/local"
GO_PATH="/usr/local/go/bin"
GO_MTA_DIR="/opt/go-mta"
QUEUE_DIR="/var/mailqueue"
BIN_MTA_SERVER="/usr/local/bin/mta-server"
BIN_MTA_QUEUE="/usr/local/bin/mta-queue"
USERS_FILE="users.txt"
CERT_FILE="/etc/ssl/certs/ssl-cert-snakeoil.pem"
KEY_FILE="/etc/ssl/private/ssl-cert-snakeoil.key"

# Install Go from official source
install_go() {
  echo "[+] Installing Go $GO_VERSION..."
  rm -rf "$GO_INSTALL_DIR/go"
  wget "$GO_URL"
  tar -C "$GO_INSTALL_DIR" -xzf "$GO_TAR"
  echo "export PATH=\$PATH:$GO_PATH" >> /etc/profile
  export PATH="$PATH:$GO_PATH"
  source /etc/profile
  rm -f "$GO_TAR"
  echo "[+] Go installed and PATH updated."
}

# Prepare directories
prepare_directories() {
  echo "[+] Preparing directories..."
  mkdir -p "$GO_MTA_DIR"
  mkdir -p "$QUEUE_DIR"
  touch "$USERS_FILE"
  chmod 600 "$USERS_FILE"
}

# Install MTA binaries
install_mta() {
  echo "[+] Copying source code..."
  cp main.go "$GO_MTA_DIR/main.go"
  cp queue.go "$GO_MTA_DIR/queue.go"
  cd "$GO_MTA_DIR"

  echo "[+] Initializing Go module..."
  "$GO_PATH/go" mod init go-mta
  "$GO_PATH/go" get github.com/emersion/go-smtp@v0.15.0

  echo "[+] Building binaries..."
  "$GO_PATH/go" build -o "$BIN_MTA_SERVER" main.go
  "$GO_PATH/go" build -o "$BIN_MTA_QUEUE" queue.go
}


# Create systemd services
create_services() {
  echo "[+] Creating systemd service files..."

  cat <<EOF > /etc/systemd/system/mta-server.service
[Unit]
Description=Go SMTP Server (MTA)
After=network.target

[Service]
ExecStart=$BIN_MTA_SERVER
WorkingDirectory=$GO_MTA_DIR
User=root
Restart=on-failure
StandardOutput=append:/var/log/mta-server.log
StandardError=append:/var/log/mta-server.err



[Install]
WantedBy=multi-user.target
EOF

  cat <<EOF > /etc/systemd/system/mta-queue.service
[Unit]
Description=Go SMTP Delivery Queue Processor
After=network.target

[Service]
ExecStart=$BIN_MTA_QUEUE
WorkingDirectory=$GO_MTA_DIR
Restart=always
RestartSec=5
User=root
StandardOutput=append:/var/log/mta-queue.log
StandardError=append:/var/log/mta-queue.err


[Install]
WantedBy=multi-user.target
EOF

  echo "[+] Reloading and enabling services..."
  systemctl daemon-reexec
  systemctl daemon-reload
  systemctl enable mta-server.service
  systemctl enable mta-queue.service
  systemctl start mta-server.service
  systemctl start mta-queue.service
}

# TLS cert reminder
tls_reminder() {
  if [[ ! -f "$CERT_FILE" || ! -f "$KEY_FILE" ]]; then
    echo "[!] TLS cert/key not found. Please place cert.pem and key.pem in /root."
  else
    echo "[+] TLS cert/key found."
  fi
}

# Main
install_go
prepare_directories
install_mta
tls_reminder
create_services

echo "[✓] MTA installation completed successfully!"
