#!/usr/bin/env bash

SERVICE_NAME=styx

# Stop, disable and remove service
systemctl stop $SERVICE_NAME
systemctl disable $SERVICE_NAME
# rm -rf /lib/systemd/system/$SERVICE_NAME.service
