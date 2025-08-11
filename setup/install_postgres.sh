#!/bin/bash
if ! command -v psql &> /dev/null; then
    echo "Installing PostgreSQL..."
    sudo apt update
    sudo apt install -y postgresql postgresql-contrib
fi

sudo systemctl enable postgresql
sudo systemctl start postgresql
