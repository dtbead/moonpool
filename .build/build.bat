sqlc generate -f "internal/db/archive_sqlc.yaml"
sqlc generate -f "internal/db/thumbnail_sqlc.yaml"

tailwindcss -c "internal/www/tailwind.config.js" -i "internal/www/web/assets/static/style.css" -o "internal/www/web/assets/static/tailwind.css" --watch

go build -C "cmd/moonpool" -a -installsuffix cgo -ldflags '-s'