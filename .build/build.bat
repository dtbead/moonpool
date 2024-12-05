sqlc generate -f "internal/db/archive_sqlc.yaml"
sqlc generate -f "internal/db/thumbnail_sqlc.yaml"

tailwindcss-windows-x64.exe -c "internal/www/tailwind.config.js" -i "internal/www/assets/static/style.css" -o "internal/www/assets/static/tailwind.css"

go build -C "cmd/moonpool" -a -installsuffix cgo -ldflags '-s'