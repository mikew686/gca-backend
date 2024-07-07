"""
Regional mapper.
Creates a file regions/NAME/solarmap.json with an array of lat longs to json files.
"""

import os
import json
import geojson
from shapely.geometry import shape, Point

NASA_ROOT = "data/nasa"
STATES_ROOT = "data/states"
STATES_GEOJSON = "data/states/us-states.json"

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

def assign_states(lat_longs):
    if not os.path.exists(STATES_GEOJSON):
        return
    with open(STATES_GEOJSON) as f:
        geo_data = geojson.load(f)
        for feature in geo_data['features']:
            polygon = shape(feature['geometry'])
            state = feature['properties']['name']
            rdat = []
            for lat, lon in lat_longs:
                point = Point(lon, lat)
                if polygon.contains(point):
                    rdat.append((lat, lon))
            mfn = os.path.join(STATES_ROOT, state+".json")
            with open(mfn, "w") as f:
                json.dump({"data":rdat, "meta":{"description":"map of solar data coordinates to state"}}, f, indent=2)
            print(f"created {mfn}")

if __name__ == "__main__":
    x = traverse_lat_long_directory(NASA_ROOT)
    # requires a state geojson file
    # example: https://raw.githubusercontent.com/PublicaMundi/MappingAPI/master/data/geojson/us-states.json
    assign_states(x)
