package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strings"
	"time"
)

const end_line_char = 10

var file_list []string
var clients []net.Conn

func main() {
	fmt.Print("\033[H\033[2J")
	file_list = make([]string, 0)
	clients = make([]net.Conn, 0)
	file, error := os.Create("ServerMessages.txt")

	if error != nil {
		fmt.Println("Found Error Creating File ServerMessages:", error)
		file.Close()
		return
	}

	file.WriteString("Messages At " + time.Now().Format("2006-01-02 15:04:05 Monday"))
	file.Close()

	go ServerService()
	var input string
	fmt.Scanln(&input)
	clients = nil
	DeleteFiles()
	fmt.Println("Closing Server At: " + time.Now().Format("2006-01-02 15:04:05 Monday"))
}

func ServerService() {

	fmt.Println("Starting Server At: " + time.Now().Format("2006-01-02 15:04:05 Monday"))
	server_connection, error := net.Listen("tcp", ":2021")

	if error != nil {
		fmt.Println("Found Error At Server Listen:", error)
		server_connection.Close()
		return
	}

	defer server_connection.Close()

	for {

		client_connection, error := server_connection.Accept()

		if error != nil {
			fmt.Println("Found Error At Server Listen:", error, "No Big Deal")
			continue
		}

		fmt.Println("\nConnected Client:", client_connection.RemoteAddr())

		clients = append(clients, client_connection)
		go HandleClient(client_connection)
	}

}

func HandleClient(client_connection net.Conn) {
	defer client_connection.Close()
	var data []byte
	buffer := make([]byte, 128)

	for {
		for {
			byte_count, error := client_connection.Read(buffer)

			if error != nil {
				if error == io.EOF {
					break
				} else {
					fmt.Println("Found Error At Client Reading:", error)
					return
				}
			}

			buffer = bytes.Trim(buffer[:byte_count], "\x00")
			data = append(data, buffer...)
			if data[len(data)-1] == end_line_char { //End of message, break then
				break
			}
		}

		if len(data) > 0 {

			if IsCommandFile(&data) {
				file_message := FileMessage(&data)
				WriteMessage(&file_message)
				fmt.Print(file_message)
				GetFile(&data)
				estatus := SendMessage(&data, client_connection)
				if !estatus {
					data = make([]byte, 0)
					break
				}
			} else {
				data_string := string(data)
				fmt.Print(data_string)
				WriteMessage(&data_string)
				estatus := SendMessage(&data, client_connection)
				if !estatus {
					data = make([]byte, 0)
					break
				}
				if data[0] == byte('|') && data[len(data)-2] == byte('|') {
					clients = Remove(clients, client_connection)
					data = make([]byte, 0)
					break
				}
			}

		} else {
			data = make([]byte, 0)
			break
		}

		data = make([]byte, 0)

	}

}

func SendMessage(data *[]byte, actual_client net.Conn) bool {
	for _, other_client := range clients {
		if other_client != actual_client {
			_, error := other_client.Write(*data)
			if error != nil {
				return false
			}
		}
	}
	return true
}

func Remove(clients []net.Conn, actual_client net.Conn) []net.Conn {

	var aux []net.Conn
	aux = make([]net.Conn, 0)
	for _, v := range clients {
		if v != actual_client {
			aux = append(aux, v)
		} else {
			fmt.Println("Removed Client:", v.RemoteAddr())
		}
	}
	return aux
}

func IsCommandFile(data *[]byte) bool {
	is_match, error := regexp.Match("file([ ]+)(([a-zA-Z0-9]*)+)[.]txt", *data)
	if error != nil {
		return false
	}
	return is_match
}

func FileMessage(data *[]byte) string {
	var information string = string(*data)
	var gettting string = ""
	for _, character := range information {

		if strings.HasSuffix(gettting, ".txt") {
			break
		}

		gettting = gettting + string(character)

	}

	return gettting

}

func WriteMessage(message *string) {
	file, error := os.OpenFile("ServerMessages.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)

	if error != nil {
		fmt.Println("Found Error Creating File ServerMessages:", error)
		file.Close()
		return
	}

	file.WriteString(*message)
	file.Close()
}

func GetFile(data *[]byte) {
	var information string = string(*data)
	var gettting string = ""
	var index int = 0
	for _, character := range information {
		index++
		if strings.HasSuffix(gettting, ".txt") {
			break
		}

		gettting = gettting + string(character)

	}

	file_name := FileName(&gettting)

	text := information[index:]

	real_file_name := "Server_" + file_name
	file, error := os.Create(real_file_name)

	if error != nil {
		fmt.Println("Found Error Creating File", file_name, ":", error)
		file.Close()
		return
	}

	file_list = append(file_list, real_file_name)
	defer file.Close()
	file.WriteString(text)

}

func FileName(value *string) string {
	list := strings.Split(*value, " ")

	if len(list) > 2 {
		return list[len(list)-1]
	}

	return ""
}

func DeleteFiles() {
	for _, v := range file_list {
		os.Remove(v)
	}

	os.Remove("ServerMessages.txt")
	file_list = nil

}
