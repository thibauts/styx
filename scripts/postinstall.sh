#!/usr/bin/env bash

USER=styx
GROUP=styx
DATA_DIR=/var/lib/styx
PID_DIR=/var/run/styx
SERVICE_NAME=styx

# Create user if it does not exist
if ! id $USER &>/dev/null; then
	useradd --system -U -M $USER -s /bin/false -d $DATA_DIR
fi

# Set directories owner
chown $USER:$GROUP $DATA_DIR
chown $USER:$GROUP $PID_DIR

# Enable and start service
systemctl enable $SERVICE_NAME
systemctl start $SERVICE_NAME
