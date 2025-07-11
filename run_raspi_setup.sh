#!/bin/bash
#
# Raspberry Pi "Enterprise-Grade" Fileserver Utility
#
# This script is a complete, deployable utility for managing an automated
# file server. It can be run interactively or in a fully unattended mode
# for automated deployments. It is idempotent and reversible.
#
# Version: 1.0.0
# Author: Gemini
#
# Enterprise Features:
#   - Full Unattended Execution via command-line flags.
#   - --uninstall flag to cleanly reverse all changes.
#   - --help flag for self-documentation.
#   - Filesystem Agnostic and resolves fstab conflicts.
#

# --- Default Configuration ---
readonly SAMBA_GROUP="sambausers"
readonly MOUNT_POINT="/mnt/external"
readonly SHARE_NAME="ExternalDrive"
readonly AUTOFS_MAP_FILE="/etc/auto.external"
readonly UNMOUNT_TIMEOUT=3

# --- Script Setup & Colors ---
set -e
set -o pipefail
readonly C_RESET='\033[0m' C_RED='\033[0;31m' C_GREEN='\033[0;32m' C_YELLOW='\033[0;33m' C_BLUE='\033[0;34m' C_BOLD='\033[1m'

# --- Helper Functions ---
function print_header() { echo -e "\n${C_BOLD}${C_BLUE}======================================================================\n $1\n======================================================================${C_RESET}"; }
function print_info() { echo -e "${C_YELLOW}[INFO] $1${C_RESET}"; }
function print_success() { echo -e "${C_GREEN}[SUCCESS] $1${C_RESET}"; }
function print_error() { echo -e "${C_RED}[ERROR] $1${C_RESET}" >&2; exit 1; }
function spinner_start() { set +m; echo -n "$1"; { while :; do for s in / - \\ \|; do echo -en "\r$1 $s"; sleep 0.2; done; done; } &; SPINNER_PID=$!; }
function spinner_stop() { { kill -9 "$SPINNER_PID" && wait; } 2>/dev/null; echo -e "\r$1 ${C_GREEN}✓${C_RESET}"; set -m; }
trap 'print_error "Script failed unexpectedly on line $LINENO."' ERR

# --- Usage and Help ---
function show_help() {
  echo "Usage: sudo $0 [OPTIONS]"
  echo
  echo "Installs and configures an automated file server. Run without options for interactive mode."
  echo
  echo "Options:"
  echo "  --user <username>       Specify the Samba user for non-interactive mode."
  echo "  --uuid <drive-uuid>       Specify the drive UUID for non-interactive mode."
  echo "  --non-interactive       Run in unattended mode. Requires --user and --uuid."
  echo "  --uninstall             Removes all configurations made by this script."
  echo "  --help                  Display this help message and exit."
}

# --- Main Logic Functions ---
function install() {
  run_checks
  install_dependencies
  select_drive
  configure_samba
  configure_autofs
  install_unmounter
}

function uninstall() {
  print_header "Uninstalling Fileserver Components"
  
  print_info "Disabling and stopping services..."
  systemctl disable --now smbd.service nmbd.service autofs.service unmounter.service &>/dev/null || true

  print_info "Removing configuration files..."
  # Surgically remove share from Samba config
  if [ -f /etc/samba/smb.conf ]; then
    sed -i "/^\[${SHARE_NAME}\]/,/^$/d" /etc/samba/smb.conf
    print_info "Removed [${SHARE_NAME}] section from smb.conf."
  fi
  # Remove autofs configs
  if [ -f /etc/auto.master ]; then
    sed -i "\|^/- ${AUTOFS_MAP_FILE}|d" /etc/auto.master
    print_info "Removed autofs master entry."
  fi
  rm -f "$AUTOFS_MAP_FILE"
  
  # Uninstall unmounter service and binary
  if [ -f /usr/local/bin/unmounter ]; then
    /usr/local/bin/unmounter -service uninstall &>/dev/null || true
    rm -f /usr/local/bin/unmounter
    print_info "Unmounter service and binary removed."
  fi
  
  # Restore backups if they exist
  if [ -f /etc/samba/smb.conf.bak ]; then mv /etc/samba/smb.conf.bak /etc/samba/smb.conf; fi
  if [ -f /etc/auto.master.bak ]; then mv /etc/auto.master.bak /etc/auto.master; fi
  if [ -f /etc/fstab.bak ]; then mv /etc/fstab.bak /etc/fstab; print_info "Restored fstab from backup."; fi

  print_info "Reloading systemd daemon..."
  systemctl daemon-reload

  print_success "Uninstall complete. A reboot is recommended to ensure all services are stopped."
}

# --- Component Functions ---
function run_checks() {
  print_header "Step 0: Pre-flight Checks"
  for cmd in git go lsblk blkid samba ntfs-3g; do
    if ! command -v "$cmd" &> /dev/null; then
      print_error "A required package is missing. Please run 'sudo apt install $cmd golang-go'."
    fi
  done
  print_success "All required commands are available."
}

function install_dependencies() {
  print_header "Step 1: Installing Dependencies"
  spinner_start "Updating package lists..."
  apt-get update &>/dev/null
  spinner_stop "Updating package lists..."
  # ntfs-3g allows mounting of NTFS drives, making the script more versatile
  spinner_start "Installing samba, autofs, git, go, and ntfs-3g..."
  apt-get install -y samba autofs git golang-go ntfs-3g &>/dev/null
  spinner_stop "Installing samba, autofs, git, go, and ntfs-3g..."
  print_success "All required software is installed."
}

function select_drive() {
  print_header "Step 2: Select & Prepare USB Drive"
  if [ "$INTERACTIVE_MODE" = true ]; then
    mapfile -t usb_devices < <(lsblk -o NAME,TRAN,SIZE,LABEL -dpn | grep "usb" | awk '{print "/dev/"$1}')
    if [ ${#usb_devices[@]} -eq 0 ]; then print_error "No USB drives detected. Please connect a drive."; fi

    echo "Please select the USB drive to use:"
    PS3="Enter the number of your drive: "
    select device in "${usb_devices[@]}"; do
      if [[ -n "$device" ]]; then
        DRIVE_DEVICE="$device"
        break
      else
        echo "Invalid selection. Try again."
      fi
    done
  else
    print_info "Using drive UUID provided: $UNATTENDED_UUID"
    DRIVE_DEVICE=$(blkid -U "$UNATTENDED_UUID" || echo "")
    if [ -z "$DRIVE_DEVICE" ]; then print_error "No device found for UUID $UNATTENDED_UUID."; fi
  fi
  
  DRIVE_UUID=$(blkid -s UUID -o value "$DRIVE_DEVICE" || echo "")
  DRIVE_FSTYPE=$(blkid -s TYPE -o value "$DRIVE_DEVICE" || echo "auto")
  if [ -z "$DRIVE_UUID" ]; then print_error "Could not determine UUID for $DRIVE_DEVICE. Is it formatted?"; fi

  if grep -q "UUID=$DRIVE_UUID" /etc/fstab; then
    print_info "Conflict detected: This drive's UUID is configured in /etc/fstab."
    if [ "$INTERACTIVE_MODE" = true ]; then
      read -p "Allow this script to disable it (comment it out) to prevent conflicts? [Y/n] " -n 1 -r; echo
      if [[ ! $REPLY =~ ^[Yy]$ && $REPLY != "" ]]; then print_error "Cannot proceed with fstab conflict. Please resolve manually."; fi
    fi
    cp /etc/fstab /etc/fstab.bak
    sed -i.bak "s|UUID=$DRIVE_UUID|# & # Disabled by fileserver script|" /etc/fstab
    print_success "Conflicting fstab entry disabled. Backup saved to /etc/fstab.bak."
  fi

  if findmnt -S "$DRIVE_DEVICE" >/dev/null; then
    umount "$DRIVE_DEVICE" || print_error "Could not unmount $DRIVE_DEVICE. Please unmount it manually."
  fi
  
  print_success "Selected $DRIVE_DEVICE | Filesystem: $DRIVE_FSTYPE | UUID: $DRIVE_UUID"
}

function configure_samba() {
  print_header "Step 3: Configuring Samba"
  
  if [ "$INTERACTIVE_MODE" = true ]; then
    read -p "Enter the username for Samba access (defaults to 'pi'): " SAMBA_USER
    SAMBA_USER=${SAMBA_USER:-pi}
    if ! id "$SAMBA_USER" &>/dev/null; then
      read -p "User '$SAMBA_USER' does not exist. Create it now? [Y/n] " -n 1 -r; echo
      if [[ $REPLY =~ ^[Yy]$ || $REPLY == "" ]]; then
        useradd -m -G sudo,${SAMBA_GROUP} "$SAMBA_USER" || true
        print_info "User '$SAMBA_USER' created. Please set a login password:"
        passwd "$SAMBA_USER"
      else
        print_error "Cannot proceed without a valid user."
      fi
    fi
  else
    SAMBA_USER="$UNATTENDED_USER"
    print_info "Using user provided: $SAMBA_USER"
    if ! id "$SAMBA_USER" &>/dev/null; then print_error "User '$SAMBA_USER' not found. Please create it before running unattended."; fi
  fi
  
  SAMBA_UID=$(id -u "$SAMBA_USER"); SAMBA_GID=$(id -g "$SAMBA_USER")

  if ! getent group "$SAMBA_GROUP" >/dev/null; then groupadd "$SAMBA_GROUP"; fi
  usermod -aG "$SAMBA_GROUP" "$SAMBA_USER"

  if ! grep -q "\[$SHARE_NAME\]" /etc/samba/smb.conf; then
    cp /etc/samba/smb.conf /etc/samba/smb.conf.bak; print_info "Backed up Samba config."
    cat << EOF >> /etc/samba/smb.conf

[${SHARE_NAME}]
path = ${MOUNT_POINT}
writeable = yes
create mask = 0775
directory mask = 0775
valid users = @${SAMBA_GROUP}
guest ok = no
EOF
  fi

  if [ "$INTERACTIVE_MODE" = true ]; then
      print_info "Please set the Samba password for '$SAMBA_USER'."
      smbpasswd -a "$SAMBA_USER"
  else
      print_info "Samba password for '$SAMBA_USER' must be set manually: sudo smbpasswd -a $SAMBA_USER"
  fi
  
  systemctl restart smbd.service nmbd.service
  print_success "Samba configured and restarted."
}

function configure_autofs() {
  print_header "Step 4: Configuring Autofs"
  mkdir -p "$MOUNT_POINT"
  
  echo "${MOUNT_POINT} -fstype=${DRIVE_FSTYPE},rw,umask=000,uid=${SAMBA_UID},gid=${SAMBA_GID} :/dev/disk/by-uuid/${DRIVE_UUID}" > "$AUTOFS_MAP_FILE"
  print_info "Created autofs map file for filesystem type '${DRIVE_FSTYPE}'."
  
  local autofs_master_line="/- ${AUTOFS_MAP_FILE} --timeout=${UNMOUNT_TIMEOUT}"
  if ! grep -Fxq "$autofs_master_line" /etc/auto.master; then
    if [ ! -f /etc/auto.master.bak ]; then cp /etc/auto.master /etc/auto.master.bak; fi
    echo "$autofs_master_line" >> /etc/auto.master
    print_info "Autofs master file updated."
  else
    print_info "Autofs master file already configured."
  fi

  systemctl restart autofs.service
  print_success "Autofs configured. Drive will unmount after ${UNMOUNT_TIMEOUT}s of inactivity."
}

function install_unmounter() {
  print_header "Step 5: Building and Installing 'unmounter' Service"
  if [ -f /usr/local/bin/unmounter ]; then
    print_info "'unmounter' binary already exists. Skipping build."
  else
    local TMP_DIR; TMP_DIR=$(mktemp -d); trap 'rm -rf "$TMP_DIR"' EXIT
    spinner_start "Cloning unmounter repo..."; git clone https://github.com/dryaf/unmounter.git "$TMP_DIR" &>/dev/null; spinner_stop "Cloning unmounter repo..."
    spinner_start "Building 'unmounter' binary..."; (cd "$TMP_DIR" && go build) &>/dev/null; spinner_stop "Building 'unmounter' binary..."
    mv "$TMP_DIR/unmounter" /usr/local/bin/; chown root:root /usr/local/bin/unmounter; chmod 755 /usr/local/bin/unmounter
  fi

  if ! id "unmounter" &>/dev/null; then useradd -r -s /bin/false unmounter; fi
  
  if [ ! -f /etc/systemd/system/unmounter.service ]; then
    print_info "Installing systemd service using unmounter's built-in feature..."
    /usr/local/bin/unmounter -service
  fi
  
  systemctl enable --now unmounter.service
  print_success "Unmounter service is enabled and running."
}

# --- Argument Parsing and Main Execution ---
if [ "$(id -u)" -ne 0 ]; then print_error "This script must be run with sudo."; fi

INTERACTIVE_MODE=true
UNATTENDED_USER=""
UNATTENDED_UUID=""

if [ "$#" -gt 0 ]; then
  while [[ "$#" -gt 0 ]]; do
    case $1 in
      --user) UNATTENDED_USER="$2"; shift ;;
      --uuid) UNATTENDED_UUID="$2"; shift ;;
      --non-interactive) INTERACTIVE_MODE=false ;;
      --uninstall) uninstall; exit 0 ;;
      --help) show_help; exit 0 ;;
      *) print_error "Unknown parameter passed: $1. Use --help for options.";;
    esac
    shift
  done
fi

if [ "$INTERACTIVE_MODE" = false ]; then
  if [ -z "$UNATTENDED_USER" ] || [ -z "$UNATTENDED_UUID" ]; then
    print_error "In non-interactive mode, --user and --uuid are required."
  fi
fi

# Run the main installation
install

print_header "Setup Complete!"
echo -e "${C_GREEN}Your Raspberry Pi is now an Enterprise-Grade automated file server.${C_RESET}"
echo -e "\n--- Summary ---\n  - Network Share: ${C_BOLD}smb://<your-pi-ip-address>/${SHARE_NAME}${C_RESET}\n  - Shared User:   ${C_BOLD}${SAMBA_USER}${C_RESET}\n  - Shared Drive:  ${C_BOLD}${DRIVE_DEVICE} (${DRIVE_FSTYPE})${C_RESET}\n  - Unmounter UI:  http://<your-pi-ip-address>:8080\n"
echo "A reboot is recommended: ${C_BOLD}sudo reboot${C_RESET}"
