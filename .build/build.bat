sqlc generate -f "internal/db/sqlc_archive.yaml"
sqlc generate -f "internal/db/sqlc_thumbnail.yaml"

go build -C "cmd/moonpool" -a -installsuffix cgo -ldflags '-s'