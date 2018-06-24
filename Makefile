.PHONY: run install redis clean

run: install
	heroku local

install: deckbuilder
	go install ./...

deckbuilder:
	go build

redis:
	heroku redis:cli -a anrdraft -c anrdraft

clean:
	rm -f deckbuilder
