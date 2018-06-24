#!/bin/bash
curl -X "GET" -H "Accept:\ application/json" -H "Content-type:\ application/x-www-form-urlencoded" https://netrunnerdb.com/api/2.0/public/cards > data/cards_input.json
curl -X "GET" -H "Accept:\ application/json" -H "Content-type:\ application/x-www-form-urlencoded" https://netrunnerdb.com/api/2.0/public/packs > data/packs_input.json
curl -X "GET" -H "Accept:\ application/json" -H "Content-type:\ application/x-www-form-urlencoded" https://netrunnerdb.com/api/2.0/public/factions > data/factions_input.json
