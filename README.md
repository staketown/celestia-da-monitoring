## Celestia Bridge Monitoring

---

Monitoring based on the following components to monitor DA node without pointing on additional local OTEL collector.

- grafana - displaying collected metrics 
- node_exporter - monitor server host (network, hardware).
- prometheus - capturing the metrics for Grafana
- loki - data source to display DA logs.
- promtail - sending logs to loki.
- alertmanager - integrated with discord webhook but could be integrated with any supported [receivers](https://prometheus.io/docs/alerting/latest/configuration/#receiver-integration-settings). 
- custom DA exporter - getting metrics (DA height, network height, peers, data received/sent etc.).

## Prerequisites

---

- Docker should be installed with sudo/root privilege. After installation needs to be logout and login again.
```bash
sudo apt update
sudo apt install -y ca-certificates curl gnupg lsb-release

sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list >/dev/null

sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

sudo usermod -aG docker $USER
```
- Celestia DA needs to be installed with systemd so logs are available in journalctl. Name of the unit file must match the official doc: https://docs.celestia.org/nodes/systemd/.
- Port 9100 should be opened as far node_exporter connected to host prometheus couldn't access to it over standard container name,
it points to node_exporter over public IP.
- Port 9999 should be opened to be able to access grafana.
- Make sure your DA node is accessible over RPC port and public IP. Example:
```bash
[RPC]
  Address = "0.0.0.0"
  Port = "26658"
  SkipAuth = false
```

## Getting started

---

### Download repository
```bash
git clone https://github.com/staketown/celestia-da-monitoring.git
cd celestia-da-monitoring
```

### Setup env variables
Setup bridge RPC, set your own **YOUR_BRIDGE_RPC_HERE**.
```bash
BRIDGE_RPC=YOUR_BRIDGE_RPC_HERE
echo "export BRIDGE_RPC=$YOUR_BRIDGE_RPC_HERE" >> $HOME/.bash_profile
source ~/.bash_profile
```
Setup webhook discord, set your own **YOUR_DISCORD_WEB_HOOK**.
```bash
DISCORD_WEB_HOOK=YOUR_DISCORD_WEB_HOOK
echo "export DISCORD_WEB_HOOK=$DISCORD_WEB_HOOK" >> $HOME/.bash_profile
source ~/.bash_profile
```
Next should be executed one of depends on which network you are going to use mocha or celestia
```bash
# for mainnet
echo "export BRIDGE_TOKEN=$(celestia bridge auth admin --p2p.network celestia)" >> $HOME/.bash_profile
source ~/.bash_profile
```
```bash
# or for testnet
echo "export BRIDGE_TOKEN=$(celestia bridge auth admin --p2p.network mocha)" >> $HOME/.bash_profile
source ~/.bash_profile
```

Setup PUBLIC_API for prometheus
```bash
cp prometheus/prometheus.yml.bak prometheus/prometheus.yml
sed -i "s/PUBLIC_IP/$(curl -4 ifconfig.me)/g" prometheus/prometheus.yml
```

Setup DISCORD_HOOK for alertmanager
```bash
cp prometheus/alert_manager/alertmanager.yml.bak prometheus/alert_manager/alertmanager.yml
sed -i "s|DISCORD_WEB_HOOK|$DISCORD_WEB_HOOK|g" prometheus/alert_manager/alertmanager.yml
```

## Start monitoring
Build exporter (should be done once or needs to upgrade)
```bash
docker compose build --no-cache
```

Start monitoring stack
```bash
docker compose up -d
```