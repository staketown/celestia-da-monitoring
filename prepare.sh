#!/bin/bash

### Setup env variables
BRIDGE_RPC=YOUR_BRIDGE_RPC_HERE
DISCORD_WEB_HOOK=YOUR_DISCORD_WEB_HOOK
PUBLIC_IP=$(curl -4 ifconfig.me)
NETWORK=celestia # set mocha if you need for testnet


echo "export BRIDGE_RPC=$YOUR_BRIDGE_RPC_HERE" >> $HOME/.bash_profile
echo "export DISCORD_WEB_HOOK=$DISCORD_WEB_HOOK" >> $HOME/.bash_profile
echo "export BRIDGE_TOKEN=$(celestia bridge auth admin --p2p.network $NETWORK)" >> $HOME/.bash_profile
source ~/.bash_profile

cp prometheus/prometheus.yml.bak prometheus/prometheus.yml
sed -i "s/PUBLIC_IP/$PUBLIC_IP/g" prometheus/prometheus.yml

cp prometheus/alert_manager/alertmanager.yml.bak prometheus/alert_manager/alertmanager.yml
sed -i "s|DISCORD_WEB_HOOK|$DISCORD_WEB_HOOK|g" prometheus/alert_manager/alertmanager.yml