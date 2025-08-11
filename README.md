#!/bin/bash
set -e
echo "Downloading and running ecommerce setup script..."
bash <(curl -sSL https://raw.githubusercontent.com/nathanialw/ecommerce/deploy/)
echo "Setup script finished."
