# Go-mi

simple Xiaomi Mija Reader for linux only
Just scan for sensor and print

## compile

x86: go build
Rpi: env GOOS=linux GOARCH=arm go build

## running
sudo ./go-mi
or
sudo ./go-mi -name XYZ
    where XYZ is the sensor name

