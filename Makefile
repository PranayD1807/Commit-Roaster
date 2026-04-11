BINARY=commit-roaster
PREFIX?=/opt/homebrew/bin

.PHONY: build install uninstall clean

build:
	go build -o $(BINARY) .

install: build
	cp $(BINARY) $(PREFIX)/$(BINARY)
	@echo "  ✅ Installed $(BINARY) to $(PREFIX)"

uninstall:
	rm -f $(PREFIX)/$(BINARY)
	@echo "  ✅ Removed $(BINARY) from $(PREFIX)"

clean:
	rm -f $(BINARY)
