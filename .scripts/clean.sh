rm -vf archive.*
rm -vf thumb.*
rm -vrf media

mkdir media
go build
./moonpool archive new
