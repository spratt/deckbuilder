#!/bin/bash
curl -X "GET" -H "Accept:\ application/json" -H "Content-type:\ application/x-www-form-urlencoded" https://netrunnerdb.com/api/2.0/public/cards > cards_input.json
