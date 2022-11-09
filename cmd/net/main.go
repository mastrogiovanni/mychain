package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"log"
	"net"
)

func server() {
	conn, err := net.ListenPacket("udp", ":5555")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	go func() {
		for {
			buffer := make([]byte, 1024)
			_, addr, err := conn.ReadFrom(buffer)
			if err != nil {
				panic(err)
			}
			reader := bytes.NewReader(buffer)
			len, err := reader.ReadByte()
			if err != nil {
				panic(err)
			}
			msg := make([]byte, len)
			_, err = reader.Read(msg)
			if err != nil {
				panic(err)
			}
			log.Printf("%s => '%s' (%d)", addr, string(buffer), len)
		}
	}()

	dst, err := net.ResolveUDPAddr("udp", "127.0.0.1:5555")
	if err != nil {
		log.Fatal(err)
	}

	msg := "supercazzola sticazzonica"

	var b bytes.Buffer
	foo := bufio.NewWriter(&b)
	err = foo.WriteByte(uint8(len(msg)))
	if err != nil {
		panic(err)
	}
	_, err = foo.Write([]byte(msg))
	if err != nil {
		panic(err)
	}
	foo.Flush()
	_, err = conn.WriteTo(b.Bytes(), dst)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Written => %s [ %s ]", msg, hex.EncodeToString(b.Bytes()))
	select {}

}

func main() {

	server()

}
