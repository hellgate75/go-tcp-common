package net

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var(
	DEFAULT_TIMEOUT time.Duration = 5 * time.Second
)

func readBytes(buffSize int, conn *tls.Conn, timeout time.Duration) ([]byte, error) {
	buf := make([]byte, buffSize)
	var buff *bytes.Buffer = bytes.NewBuffer([]byte{})
	var n int = 0
	var err error
	start := time.Now()
	for n <= 0 && int64(time.Now().Sub(start)) <= int64(timeout) {
		n, err = conn.Read(buf)
		if err != nil {
			return nil, err
		}
		if n > 0 {
			_, err = buff.Write(buf[:n])
			if err != nil {
				return []byte{}, err
			}
		}
	}
	buff = bytes.NewBufferString(strings.TrimSpace(string(buff.Bytes())))
	return buff.Bytes(), nil
}

func readString(buffSize int, conn *tls.Conn, timeout time.Duration) (string, error) {
	buf := make([]byte, buffSize)
	var buff *bytes.Buffer = bytes.NewBuffer([]byte{})
	var n int = 0
	var err error
	start := time.Now()
	for n <= 0 && int64(time.Now().Sub(start)) <= int64(timeout) {
		n, err = conn.Read(buf)
		if err != nil {
			return "", err
		}
		if n > 0 {
			_, err = buff.Write(buf[:n])
			if err != nil {
				return "", err
			}
		}
		//fmt.Printf("Reading data - received: %v, time passed: %s, timeout: %s\n", (n > 0), time.Now().Sub(start).String(), timeout.String())
	}
	return strings.TrimSpace(buff.String()), nil
}

func ReadStringTimeout(conn *tls.Conn, timeout time.Duration) (string, error) {
	return readString(2048, conn, timeout)
}

func ReadString(conn *tls.Conn) (string, error) {
	return ReadStringTimeout(conn, DEFAULT_TIMEOUT)
}

func ReadStringBufferTimeout(buffSize int, conn *tls.Conn, timeout time.Duration) (string, error) {
	return readString(buffSize, conn, timeout)
}

func ReadStringBuffer(buffSize int, conn *tls.Conn) (string, error) {
	return ReadStringBufferTimeout(buffSize, conn, DEFAULT_TIMEOUT)
}

func ReadTimeout(conn *tls.Conn, timeout time.Duration) ([]byte, error) {
	value, errX := readString(2048, conn, timeout)
	if errX != nil {
		return nil, errX
	}
	size, errY := strconv.Atoi(value)
	if errY != nil {
		return nil, errY
	}
	return readBytes(size, conn, timeout)
}

func Read(conn *tls.Conn) ([]byte, error) {
	return ReadTimeout(conn, DEFAULT_TIMEOUT)
}

func writeString(value string, conn *tls.Conn) (int, error) {
	return conn.Write([]byte(value))

}

func writeBytes(value []byte, conn *tls.Conn) (int, error) {
	return conn.Write(value)

}

func WriteString(value string, conn *tls.Conn) (int, error) {
	value = strings.TrimSpace(value)
	return writeString(value, conn)
}

func Write(value []byte, conn *tls.Conn) (int, error) {
	value = []byte(strings.TrimSpace(fmt.Sprintf("%s", string(value))))
	n, err := writeString(fmt.Sprintf("%v", len(value)), conn)
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, errors.New(fmt.Sprintf("No bytes written as frame-size"))
	}
	return writeBytes(value, conn)

}
