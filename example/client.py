#!/usr/bin/env python3

import requests
from datetime import datetime as dt
import time

def get_data(i, url):
    """perform a GET request to the given URL and return the response as a dictionary
    This is only to demonstrate "doing something" so we can send some stats to the statwriter
    """

    # the response dictionary
    res = {
        "value": [],
        "time": None,
        "id": None,
        "extra": []
    }

     # this is the time in ISO 8601 format + Z
    t = dt.now().strftime("%Y-%m-%dT%H:%M:%S")
    t += "Z"

    # try to parse the response and fill the dictionary
    try:
        response = requests.get(url)
        data = response.json()

        # i happen to know in this case that the response is a dictionary with a key 'data'
        out = data['data']

        elapsed = response.elapsed.total_seconds()

        res['value'] = [elapsed]
        res['time'] = t
        res['id'] = out
        res['extra'] = [{"iteration": i, "url": url}]
    except Exception as e:
        # dont blow up if something goes wrong, just inform the user
        print(e)
    return res

def post_to_statwriter(data, url):
    response = requests.post(url, json=data)
    print(response.status_code)

def main():
    resource_uri = "https://namer.nullferatu.com"
    statwriter_uri = "http://cobra.nullferatu.com:20080/namer-stats"
    for i in range(0, 10):
        data = get_data(i, resource_uri)
        post_to_statwriter(data, statwriter_uri)
        time.sleep(2)

if __name__ == "__main__":
    main()
