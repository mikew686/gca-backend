"""
Goes through available regions, and updates their MOER files with the following format:
Hourly MOER average for 2023:
{
 "data":
 {
  "lbs_co2_per_mwh": {
    "2023010100": 0,
    ...
  }
 }
}
"""

import os
import json
from collections import defaultdict

def update_moer_files(region):
    """
    Reorganize the MOER data to format as the NASA power data return.
    """
    folder_path = os.path.join("data", "regions", region)
    if not os.path.exists(folder_path):
        print(f"region data not found: {region}")
        return
    data = defaultdict(lambda: defaultdict(list))
    prefix = f"{region}_2023"
    for filename in os.listdir(folder_path):
        if filename.startswith(prefix) and filename.endswith('.json'):
            filepath = os.path.join(folder_path, filename)
            with open(filepath, 'r') as f:
                dat = json.load(f)
                for row in dat["data"]:
                    timestamp, moer = row["point_time"], float(row["value"])
                    date, time = timestamp.split('T')
                    hour = time.split(':')[0]
                    data[date.replace("-", "")][hour].append(moer)
    outdat = {}
    for date, hours in data.items():
        for hour, values in hours.items():
            average_value = round(sum(values) / len(values))
            outdat[date+hour] = average_value
    filedat = {"data": {"lbs_co2_per_mwh": outdat}}
    outpath = os.path.join("data", "regions", region, "moer.json")
    with open(outpath, "w") as f:
        json.dump(filedat, f, indent=1)
    print(f"created {outpath}")

if __name__ == "__main__":
    with open('data/ba_maps.json', 'r') as f:
        data = json.load(f)
        regions = [feature['properties']['region'] for feature in data['features']]
        for region in regions:
            update_moer_files(region)
