# Build variables
APP_NAME = Letta
BINARY = letta
APP_BUNDLE = $(APP_NAME).app
DMG_NAME = $(APP_NAME).dmg
INSTRUCTIONS_FILE = Инструкция.txt

.PHONY: build run test clean app dmg help

# Default target
help:
	@echo "Available targets:"
	@echo "  build    - Build the binary"
	@echo "  run      - Run the application"
	@echo "  test     - Run tests"
	@echo "  app      - Create .app bundle"
	@echo "  dmg      - Create DMG installer with instructions"
	@echo "  clean    - Clean all generated files"

# Build binary
build:
	go build -o $(BINARY) ./cmd/letta

# Run application
run:
	go run ./cmd/letta

# Run tests
test:
	go test ./...

# Create Info.plist file
Info.plist:
	@echo '<?xml version="1.0" encoding="UTF-8"?>' > $@
	@echo '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' >> $@
	@echo '<plist version="1.0">' >> $@
	@echo '<dict>' >> $@
	@echo '    <key>CFBundleExecutable</key>' >> $@
	@echo '    <string>$(BINARY)</string>' >> $@
	@echo '    <key>CFBundleIdentifier</key>' >> $@
	@echo '    <string>com.$(APP_NAME).app</string>' >> $@
	@echo '    <key>CFBundleName</key>' >> $@
	@echo '    <string>$(APP_NAME)</string>' >> $@
	@echo '    <key>CFBundleDisplayName</key>' >> $@
	@echo '    <string>$(APP_NAME)</string>' >> $@
	@echo '    <key>CFBundleVersion</key>' >> $@
	@echo '    <string>1.0</string>' >> $@
	@echo '    <key>CFBundleShortVersionString</key>' >> $@
	@echo '    <string>1.0</string>' >> $@
	@echo '    <key>CFBundlePackageType</key>' >> $@
	@echo '    <string>APPL</string>' >> $@
	@echo '    <key>NSHighResolutionCapable</key>' >> $@
	@echo '    <true/>' >> $@
	@echo '</dict>' >> $@
	@echo '</plist>' >> $@
	@echo "Created $@"

# Create instructions file (UTF-8 encoded)
$(INSTRUCTIONS_FILE):
	@echo "📦 Установка $(APP_NAME)" > $@
	@echo "" >> $@
	@echo "Инструкция находится по адресу:" >> $@
	@echo "" >> $@
	@echo "    https://github.com/azamuray/letta/wiki/instruction" >> $@
	@echo "" >> $@

# Create .app bundle (depends on Info.plist)
app: build Info.plist
	rm -rf $(APP_BUNDLE)
	mkdir -p $(APP_BUNDLE)/Contents/MacOS
	cp $(BINARY) $(APP_BUNDLE)/Contents/MacOS/
	cp Info.plist $(APP_BUNDLE)/Contents/
	chmod +x $(APP_BUNDLE)/Contents/MacOS/$(BINARY)
	@echo "Created $(APP_BUNDLE)"

# Create DMG installer with instructions
dmg: app $(INSTRUCTIONS_FILE)
	rm -f $(DMG_NAME)
	
	# Создаем временную папку с содержимым для DMG
	mkdir -p dist-for-dmg
	cp -R $(APP_BUNDLE) dist-for-dmg/
	cp $(INSTRUCTIONS_FILE) dist-for-dmg/
	
	create-dmg \
	  --volname "$(APP_NAME)" \
	  --window-pos 200 120 \
	  --window-size 800 500 \
	  --icon-size 90 \
	  --icon "$(APP_BUNDLE)" 200 180 \
	  --icon "$(INSTRUCTIONS_FILE)" 400 180 \
	  --hide-extension "$(APP_BUNDLE)" \
	  --text-size 14 \
	  --app-drop-link 600 185 \
	  "$(DMG_NAME)" \
	  "dist-for-dmg/"
	  
	@echo ""
	@echo "✅ Создан $(DMG_NAME)"
	@echo "📁 Внутри DMG:"
	@echo "   • $(APP_BUNDLE) - ваше приложение"
	@echo "   • $(INSTRUCTIONS_FILE) - инструкция по установке"
	
	# Удаляем временную папку
	rm -rf dist-for-dmg

# Clean all generated files
clean:
	rm -rf $(APP_BUNDLE) $(BINARY) $(DMG_NAME) Info.plist $(INSTRUCTIONS_FILE) dist-for-dmg