cd "E:\Programming\go\src\github.com\dtbead\moonpool\db"
sqlc generate

cd "E:\Programming\go\src\github.com\dtbead\moonpool"
"E:\Programming\go\src\github.com\dtbead\moonpool\.build\tailwindcss-windows-x64.exe" -i server/www/assets/static/style.css -o server/www/assets/static/tailwind.css

cd "E:\Programming\go\src\github.com\dtbead\moonpool\cmd\moonpool"
go build