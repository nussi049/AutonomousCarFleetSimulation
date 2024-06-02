#!/bin/bash

# Parameter prüfen
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <number_of_cars> <max_value>"
    exit 1
fi

# Anzahl der Autos und maximaler Zufallswert
num_cars=$1
max_value=$2

# Farben
colors=("Rot" "Grün" "Blau" "Cyan" "Magenta" "Orange" "Pink" "Lila" "Braun" "Schwarz")

# Server starten
echo "Starting server..."
go run coordinator/cmd/main.go &
server_pid=$!
echo "Server started with PID $server_pid"

# 2 Sekunden warten
sleep 2

# Autos starten
for i in $(seq 1 $num_cars)
do
    port=$((50001 + i))
    color=${colors[$(( (i - 1) % ${#colors[@]} ))]}
    x=$((RANDOM % max_value))
    y=$((RANDOM % max_value))
    echo "Starting car $i on port $port with color $color, x=$x, y=$y..."
    go run carclient/cmd/main.go --port=$port --color=$color --x=$x --y=$y &
done

# Auf Eingabe warten, um die Skriptausführung zu stoppen
read -p "Press any key to stop all processes..."

# Prozesse stoppen
echo "Stopping server..."
kill $server_pid

echo "Stopping cars..."
pkill -f "go run main.go"

echo "All processes stopped."
