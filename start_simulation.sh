#!/bin/bash

if [ "$#" -ne 3 ]; then
    echo "Usage: $0 <number_of_cars> <max_value> <advancedDrive>"
    exit 1
fi

num_cars=$1
max_value=$2
advanced_drive=$3

colors=("Rot" "Grün" "Blau" "Cyan" "Magenta" "Orange" "Pink" "Lila" "Braun" "Schwarz")

echo "Starting server..."
go run coordinator/cmd/main.go &
server_pid=$!
echo "Server started with PID $server_pid"

sleep 2

for i in $(seq 1 $num_cars)
do
    port=$((50001 + i))
    color=${colors[$(( (i - 1) % ${#colors[@]} ))]}
    x=$((RANDOM % max_value))
    y=$((RANDOM % max_value))
    echo "Starting car $i on port $port with color $color, x=$x, y=$y, advancedDrive=$advanced_drive..."
    go run carclient/cmd/main.go --port=$port --color=$color --x=$x --y=$y --advancedDrive=$advanced_drive &
done

read -p "Press any key to stop all processes..."

echo "Stopping server..."
kill $server_pid

echo "Stopping cars..."
pkill -f "go run main.go"

echo "All processes stopped."