.PHONY: run install

run: install
	heroku local

install: deckbuilder
	go install ./...

deckbuilder:
	go build

clean:
	rm -f deckbuilder
