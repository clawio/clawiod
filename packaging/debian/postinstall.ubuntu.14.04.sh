#!/bin/sh

set -e

APP_NAME="clawiod"
APP_USER="clawiod"
APP_GROUP="clawiod"
CORE_CONFIG="/opt/${APP_NAME}/config.core.json"
FILE_AUTH_CONFIG="/opt/${APP_NAME}/conf.d/config.fileauth.json"


sudo cp ${CORE_CONFIG} /etc/${APP_NAME}/config.core.json
sudo cp ${FILE_AUTH_CONFIG} /etc/${APP_NAME}/conf.d/config.fileauth.json
