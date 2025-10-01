package main

import (
    "bufio"
    "crypto/tls"
    "flag"
    "fmt"
    "io"
    "log"
    "net"
    "net/url"
    "os"
    "strings"
)

func main() {
    proxyURL := flag.String("proxy", "", "Proxy URL (e.g. http://localhost:8080/proxy/w-1234)")
    token := flag.String("token", "", "JWT token for authentication")
    flag.Parse()

    if *proxyURL == "" || *token == "" {
        flag.Usage()
        os.Exit(1)
    }

    u, err := url.Parse(*proxyURL)
    if err != nil {
        log.Fatalf("invalid proxy url: %v", err)
    }

    host := u.Host
    if !strings.Contains(host, ":") {
        if u.Scheme == "https" {
            host = host + ":443"
        } else {
            host = host + ":80"
        }
    }

    var conn net.Conn
    if u.Scheme == "https" {
        conn, err = tls.Dial("tcp", host, &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: true})
    } else {
        conn, err = net.Dial("tcp", host)
    }
    if err != nil {
        log.Fatalf("failed to connect to proxy host: %v", err)
    }
    defer conn.Close()

    hostHeader := u.Host
    requestLine := fmt.Sprintf("CONNECT %s HTTP/1.1\r\n", pathWithQuery(u))
    headers := fmt.Sprintf("Host: %s\r\nAuthorization: Bearer %s\r\n\r\n", hostHeader, *token)

    if _, err := io.WriteString(conn, requestLine+headers); err != nil {
        log.Fatalf("failed to write CONNECT request: %v", err)
    }

    br := bufio.NewReader(conn)
    statusLine, err := br.ReadString('\n')
    if err != nil {
        log.Fatalf("failed to read proxy response: %v", err)
    }
    if !strings.Contains(statusLine, "200") {
        // drain rest for logging
        body, _ := io.ReadAll(br)
        log.Fatalf("proxy CONNECT failed: %s%s", statusLine, body)
    }

    // skip remaining response headers
    for {
        line, err := br.ReadString('\n')
        if err != nil {
            log.Fatalf("failed to read proxy headers: %v", err)
        }
        if line == "\r\n" {
            break
        }
    }

    // Pipe stdin/out to the proxy connection
    go func() {
        _, _ = io.Copy(conn, os.Stdin)
    }()
    _, _ = io.Copy(os.Stdout, conn)
}

func pathWithQuery(u *url.URL) string {
    path := u.RequestURI()
    if path == "" {
        path = "/"
    }
    return path
}
