package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
)

const port = ":8080"
const targetPort = ":8090"

func main() {
	cert, err := tls.LoadX509KeyPair("cod.uno.crt.pem", "cod.uno.key.pem")
	if err != nil {
		panic(err)
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	listener, err := tls.Listen("tcp", port, config)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Listening...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
		}
		go proxy(conn)
	}
}

func proxy(down net.Conn) {
	defer down.Close()
	log.Printf("Handling this!")
	targetIp, err := ip()
	if err != nil {
		log.Print("ip: " + err.Error())
		return
	}
	target := targetIp + targetPort
	log.Printf("Piping to %q", target)

	up, err := net.Dial("tcp", target)
	if err != nil {
		log.Print(err)
		return
	}
	defer up.Close()

	log.Printf("Piping streams...")
	go copy(up, down)
	copy(down, up)
}

func copy(w io.Writer, r io.Reader) {
	n, err := io.Copy(io.MultiWriter(w, os.Stderr), r)
	if err != nil {
		log.Printf("copy: %s", err)
		return
	}
	log.Printf("copy: %d", n)
}

func ip() (string, error) {
	r, err := http.Get("https://api.cod.uno/ip")
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		return "", fmt.Errorf("got %d", r.StatusCode)
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
