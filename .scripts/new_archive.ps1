# new_archive deletes any existing moonpool related database files and creates a new blank archive.

Set-Location .\cmd\moonpool

Remove-Item -Recursive profile
Remove-Item -Recursive media
Remove-Item archive.sqlite3 
Remove-Item thumb.db
go build && ./moonpool.exe archive new