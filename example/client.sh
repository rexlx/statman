#!/bin/bash

url="http://localhost:20080/lan-access"

alive=0
res=$(fping -aAqg 192.168.86.0/24)
now=$(date +"%Y-%m-%dT%H:%M:%S")

for i in $res;do alive=$((alive+1));do
extra=$(echo $res | tr '\n' ',' | sed 's/,$//')

genPostData() {
    cat <<EOF
{
    "time": "${now}Z",
    "value": [${alive}],
    "id": "1",
    "extra": ["${extra}"]
}
EOF
}

data=$(genPostData)

echo "$data"
echo "POST to $url"

curl -i \
-H "Accept: application/json" \
-H "Content-Type:application/json" \
-X POST "$url" --data "$data"

