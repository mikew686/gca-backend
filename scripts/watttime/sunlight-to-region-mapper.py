"""
Regional mapper.
Creates a file regions/NAME/solarmap.json with an array of lat longs to json files.
"""

import os
import json
import geojson
from shapely.geometry import shape, Point

NASA_ROOT = "data/nasa"
REGIONS_ROOT = "data/regions"
REGION_GEOJSON = "data/ba_maps.json"
SOLAR_MAPFILE = "solarmap.json"

def traverse_lat_long_directory(base_dir):
    lat_lon_list = []
    for root, dirs, files in os.walk(base_dir):
        if root == base_dir:
            continue
        for file in files:
            if file.endswith(".json"):
                lat = os.path.basename(root)
                lon = file.replace(".json", "")
                lat_lon_list.append((float(lat), float(lon)))
    return lat_lon_list

with open(REGION_GEOJSON) as f:
    geo_data = geojson.load(f)

def assign_regions(lat_longs):
    for feature in geo_data['features']:
        polygon = shape(feature['geometry'])
        region = feature['properties']['region']
        if not os.path.exists(os.path.join(REGIONS_ROOT, region)):
            continue
        rdat = []
        for lat, lon in lat_longs:
            point = Point(lon, lat)
            if polygon.contains(point):
                rdat.append((lat, lon))
        mfn = os.path.join(REGIONS_ROOT, region, SOLAR_MAPFILE)
        with open(mfn, "w") as f:
            json.dump({"data":rdat, "meta":{"description":"map of solar data coordinates to region"}}, f, indent=2)
        print(f"created {mfn}")

if __name__ == "__main__":
    x = traverse_lat_long_directory(NASA_ROOT)
    assign_regions(x)
