$Amount = 500
$Range = 250

for ($i = 0; $i -lt $Amount; $i++) {
    $randomNumber = Get-Random -Maximum $Range -Minimum 1
    Invoke-WebRequest -URI "http://127.0.0.1:9996/post/entry/$randomNumber"
}
