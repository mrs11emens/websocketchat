package main

import (
	"fmt"
	"net/http"
	"wschat/server"
)

var addr string = "192.168.10.25:8083"

func main() {
	http.HandleFunc("/ws", server.HandleConnection)
	
	// Форматирование строки с переменной
	fmt.Printf("WebSocket server started at ws://%s/ws\n", addr)
	
	// Путь к статическим файлам
	staticDir := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", http.StripPrefix("/", staticDir))
	
	// Запуск сервера
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}