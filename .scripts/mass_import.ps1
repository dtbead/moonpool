Get-ChildItem -LiteralPath "E:\Hydrus Network SFW\import" | ForEach-Object {
    if ($_.Extension -eq ".png") {
        $filename = $_.FullName
        $J = (curl -F file=@"$filename" "http://127.0.0.1:5878/post/upload") | ConvertFrom-JSON

        $random = -join ((48..57) + (97..122) | Get-Random -Count 12 | % {[char]$_})
        if($null-ne $J.id) {
            $whatever = '["' + $random + '"]' 
            curl "http://127.0.0.1:5878/post/set_tags/$($J.id)" -H "Content-Type: application/json" --data $whatever
        }
    }
}