#!/bin/bash
set -e
echo "Downloading and running ecommerce setup script..."
bash <(curl -sSL https://raw.githubusercontent.com/nathanialw/ecommerce/deploy/)
echo "Setup script finished."

git add .
git commit -m "updated"
git push
git tag v0.0.16
git push origin v0.0.16


go list -m -versions github.com/nathanialw/ecommerce