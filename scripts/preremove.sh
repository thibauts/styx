#!/usr/bin/env bash

SERVICE_NAME=styx

# Stop and disable service
systemctl stop $SERVICE_NAME
systemctl disable $SERVICE_NAME
