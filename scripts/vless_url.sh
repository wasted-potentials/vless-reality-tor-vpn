
#!/usr/bin/env bash
# helper to print VLESS URL from config and first user

CFG=${1:-example/config.yaml}
USERS_JSON=$(grep -oP '(?<=path: ).*' "$CFG" | tr -d '"')

if [ ! -f "$USERS_JSON" ]; then
  echo "Users file $USERS_JSON not found"
  exit 1
fi

UUID=$(jq -r '.[0].uuid' "$USERS_JSON")
DEST=$(grep -oP '(?<=dest: ).*' "$CFG" | tr -d '"')
SNI=$(grep -oP '(?<=server_names: \[).*?(?=\])' "$CFG" | tr -d '"' | cut -d',' -f1 | xargs)
HOSTPORT=$(grep -oP '(?<=listen: ).*' "$CFG" | tr -d '"' )

HOST=$(hostname -I | awk '{print $1}')
PORT=$(echo $HOSTPORT | awk -F: '{print $2}')
SID=$(grep -oP '(?<=short_ids: \[).*?(?=\])' "$CFG" | tr -d '"' | cut -d',' -f1 | xargs)

echo "vless://$UUID@$HOST:$PORT?encryption=none&flow=xtls-rprx-vision&security=reality&sni=$SNI&sid=$SID&type=tcp#vpnd"
