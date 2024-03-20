#!/bin/bash

CATEGORY="$1"
if [ -z "$CATEGORY" ]; then
    echo "Usage: $0 <armor_type>"
    echo "Armor types: helmet, cuirass, gloves, boots, cape, amulet, ring, shield"
    exit 1
fi



recsel -C -e "armor_type = '${CATEGORY}'" -P internal_name ./files/assets/data/items/armor.rec