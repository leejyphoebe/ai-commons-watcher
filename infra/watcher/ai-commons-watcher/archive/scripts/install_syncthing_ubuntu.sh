#!/usr/bin/env bash
set -e

echo "Installing Syncthing (Ubuntu/Debian)..."

sudo apt-get update
sudo apt-get install -y curl apt-transport-https gnupg

curl -s https://syncthing.net/release-key.txt | \
  sudo gpg --dearmor -o /usr/share/keyrings/syncthing-archive-keyring.gpg

echo "deb [signed-by=/usr/share/keyrings/syncthing-archive-keyring.gpg] https://apt.syncthing.net/ syncthing stable" | \
  sudo tee /etc/apt/sources.list.d/syncthing.list

sudo apt-get update
sudo apt-get install -y syncthing

echo
echo "Syncthing installed."
echo "To start it for this user, run:"
echo "   syncthing serve --no-browser &"
