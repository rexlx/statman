#!/usr/bin/env python3

import requests
from datetime import datetime as dt
import time

def get_data(i, url):
    res = {
        "value": [],
        "time": None,
        "id": None,
        "extra": []
    }
    response = requests.get(url)
    try:
        t = dt.now().strftime("%Y-%m-%dT%H:%M:%S")
        t += "Z"
        data = response.json()
        out = data['data']
        elapsed = response.elapsed.total_seconds()
        res['value'] = [elapsed]
        res['time'] = t
        res['id'] = out
        res['extra'] = [{"iteration": i, "url": url}]
    except Exception as e:
        print(e)
    return res

def post_to_statwriter(data, url):
    response = requests.post(url, json=data)
    print(response.status_code)

def main():
    resource_uri = "https://namer.nullferatu.com"
    statwriter_uri = "http://cobra.nullferatu.com:20080/namer-stats"
    for i in range(0, 1000):
        data = get_data(i, resource_uri)
        post_to_statwriter(data, statwriter_uri)
        time.sleep(2)

if __name__ == "__main__":
    main()
