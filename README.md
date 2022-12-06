# ServerStatus

以针会友

## Install
```shell
wget -q https://www.github.com/DGHeroin/ServerStatus/releases/latest/download/ServerStatus-linux-amd64 \
  -O /usr/local/bin/ServerStatus && \
  chmod +x /usr/local/bin/ServerStatus
```

### server systemd
```shell
[Unit]
Description=ServerStatus Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/usr/local/bin/
ExecStart=ServerStatus server --addr ":16808" --authAgent token_of_agent --authView token_of_view_api
Restart=always

[Install]
WantedBy=multi-user.target
```

### agent systemd
```shell
[Unit]
Description=ServerStatus Agent
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/usr/local/bin/
ExecStart=ServerStatus agent --addr "http://127.0.0.1:16808" --auth token_of_agent --s vps-node-name
Restart=always

[Install]
WantedBy=multi-user.target
```

### view by curl
```shell
curl -H "Authorization: token_of_view_api" http://127.0.0.1:16808/api/view/status

curl -H "Authorization: token_of_view_api" http://127.0.0.1:16808/api/view/status?txt=1
```
