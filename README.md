# chat-app
go run main.go
docker run -p 8080:8080 chat-app:latest
docker build --network host -t chat-app:latest .
docker run --network host chat-app:latest
sudo netstat -tuln | grep 5672