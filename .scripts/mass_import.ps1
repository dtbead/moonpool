Get-ChildItem -LiteralPath "E:\Hydrus Network SFW\import" | ForEach-Object {
    if ($_.Extension -eq ".png") {
        $filename = $_.FullName
        curl -F file=@"$filename" "http://127.0.0.1:5878/post/upload"
    }
}