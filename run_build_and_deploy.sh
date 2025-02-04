// ==== File: run_build_and_deploy.sh ====
#!/bin/bash

# Load environment variables from a .env file if it exists
if [ -f ".env" ]; then
    echo "Loading environment variables from .env file..."
    source .env
else
    echo ".env file not found. Proceeding with existing environment variables or prompts."
fi

# Set or prompt for values
REMOTE_USER="${REMOTE_USER:-}"
REMOTE_IP="${REMOTE_IP:-}"
REMOTE_HOME="${REMOTE_HOME:-/home/$REMOTE_USER}" # Default to /home/$REMOTE_USER if not specified

if [ -z "$REMOTE_USER" ]; then
  read -p "Enter remote username: " REMOTE_USER
fi
if [ -z "$REMOTE_IP" ]; then
  read -p "Enter remote IP address: " REMOTE_IP
fi
if [ -z "$REMOTE_HOME" ]; then
  read -p "Enter remote home directory [/home/$REMOTE_USER]: " REMOTE_HOME
  REMOTE_HOME=${REMOTE_HOME:-/home/$REMOTE_USER}
fi

# Confirm provided or entered details
echo "Using the following configuration:"
echo "Remote user: $REMOTE_USER"
echo "Remote IP: $REMOTE_IP"
echo "Remote home directory: $REMOTE_HOME"

# Function to execute remote commands using expect
execute_remote_commands_with_expect() {
    local password="$1"
    local commands="$2"
    expect -c "
        set timeout 10
        spawn ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_IP \"$commands\"
        expect \"*?assword:*\"
        send \"$password\r\"
        interact
    "
}

# Read the password securely
read -s -p "Enter remote password for $REMOTE_USER: " REMOTE_PASSWORD
echo

# Ensure expect is installed on the remote machine
echo "Ensuring expect is installed on the remote machine..."
ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_IP "if ! command -v expect &> /dev/null; then sudo apt-get update && sudo apt-get install -y expect; fi" << EOF
$REMOTE_PASSWORD
EOF

# Proceed with the script
# Step 1: Set up Go environment for cross-compilation
export GOOS=linux
export GOARCH=arm64
export CGO_ENABLED=0

# Step 2: Build the Go project
echo "Building the project..."
go build -o unmounter  -ldflags "-X 'main.username=${AUTH_USER}' -X 'main.password=${AUTH_PASS}'" .
if [ $? -ne 0 ]; then
    echo "Build failed. Exiting."
    exit 1
fi
echo "Build succeeded."

# Step 3: uninstall and stop existing service
echo "Executing uninstall and stop commands on the remote server..."
execute_remote_commands_with_expect "$REMOTE_PASSWORD" "sudo $REMOTE_HOME/unmounter uninstall; sudo $REMOTE_HOME/unmounter stop"

# Step 4: SCP the binary to the target machine
echo "Copying the binary to $REMOTE_IP..."
expect -c "
    set timeout 10
    spawn scp unmounter $REMOTE_USER@$REMOTE_IP:$REMOTE_HOME/
    expect \"*?assword:*\"
    send \"$REMOTE_PASSWORD\r\"
    expect eof
"
if [ $? -ne 0 ]; then
    echo "SCP failed. Exiting."
    exit 1
fi
echo "Copy succeeded."

# Step 5: Execute install and start service
echo "Executing install and start commands on the remote server..."
execute_remote_commands_with_expect "$REMOTE_PASSWORD" "sudo chmod +x $REMOTE_HOME/unmounter; sudo $REMOTE_HOME/unmounter install; sudo $REMOTE_HOME/unmounter start"

if [ $? -ne 0 ]; then
    echo "Remote command execution failed."
else
    echo "Remote commands executed successfully."
fi
