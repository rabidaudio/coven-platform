set -e
ip=$(ifconfig | grep "inet " | grep -v 127.0.0.1 | awk '{ print $2 }')
echo "API_URL=http://${ip}:8080"
flutter run --dart-define=API_URL=http://${ip}:8080
