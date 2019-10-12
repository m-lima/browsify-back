package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

var (
	smtpLogStd = log.New(os.Stdout, "[smtp] ", log.Ldate|log.Ltime)
	smtpLogErr = log.New(os.Stderr, "ERROR [smtp] ", log.Ldate|log.Ltime)
)

func transfer(conn net.Conn, message string) error {
	buffer := make([]byte, 256)

	if len(message) > 0 {
		bytes, err := conn.Write([]byte(message + "\x0d\x0a"))
		if err != nil {
			return err
		}
		if bytes < len(message)+2 {
			return fmt.Errorf("Was not able to write full payload. Sent %d out of %d", bytes, len(message)+2)
		}
		smtpLogStd.Println(">", message)
	}

	bytes, err := conn.Read(buffer)
	if err != nil {
		return err
	}
	if bytes < 1 {
		return fmt.Errorf("Got an empty response from SMTP server")
	}
	smtpLogStd.Print("< ", strings.ReplaceAll(string(buffer), "\x0d\x0a", ""))

	return nil
}

func Send(user *User, payload string) {
	conn, err := net.Dial("tcp", config.Smtp.Server)
	if err != nil {
		smtpLogErr.Println(err)
		return
	}
	defer conn.Close()

	err = transfer(conn, "")
	if err != nil {
		smtpLogErr.Println(err)
		return
	}

	err = transfer(conn, fmt.Sprintf("HELO %s", config.Smtp.Identity))
	if err != nil {
		smtpLogErr.Println(err)
		return
	}

	err = transfer(conn, fmt.Sprintf("MAIL FROM: <%s>", config.Smtp.From))
	if err != nil {
		smtpLogErr.Println(err)
		return
	}

	err = transfer(conn, fmt.Sprintf("RCPT TO: <%s>", config.Smtp.To))
	if err != nil {
		smtpLogErr.Println(err)
		return
	}

	err = transfer(conn, "DATA")
	if err != nil {
		smtpLogErr.Println(err)
		return
	}

	err = transfer(conn, fmt.Sprintf("From: %s\x0d\x0aTo: %s\x0d\x0aSubject: Access log for %s\x0d\x0a\x0d\x0a%s\x0d\x0a.", config.Smtp.From, config.Smtp.To, user.Email, payload))
	if err != nil {
		smtpLogErr.Println(err)
		return
	}

	err = transfer(conn, "QUIT")
	if err != nil {
		smtpLogErr.Println(err)
		return
	}
}
