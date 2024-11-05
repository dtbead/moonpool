sqlc generate -f "internal/db/sqlc_archive.yaml"
sqlc generate -f "internal/db/sqlc_thumbnail.yaml"

tailwindcss-windows-x64.exe -c "internal/www/tailwind.config.js" -i "internal/www/assets/static/style.css" -o "internal/www/assets/static/tailwind.css"

go build -C "cmd/moonpool" -a -installsuffix cgo -ldflags '-s'