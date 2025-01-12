package tcp

import (
	"io"
	"net"
	"testing"
)

func TestDial(t *testing.T) {
	// Create a listener on a random port
	listener, err := net.Listen("tcp", "127.0.0.1:")

	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		defer func() { done <- struct{}{} }()

		for {
			conn, err := listener.Accept()
			if err != nil {
				t.Log(err)
				return
			}

			go func(c net.Conn) {
				defer func() {
					c.Close()
					done <- struct{}{}
				}()

				buf := make([]byte, 1024)
				for {
					n, err := c.Read(buf)
					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}
						t.Logf("received: %e", err)
						return
					}
					t.Logf("received: %q", buf[:n])
				}
			}(conn)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	conn.Close()
	<-done
	listener.Close()
	<-done
}

// func TestListener(t *testing.T) {
// 	listener, err := net.Listen("tcp", "127.0.0.1:0")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	defer func() { _ = listener.Close() }()

// 	t.Logf("bound to %q", listener.Addr())

// 	for {
// 		conn, err := listener.Accept()
// 		if err != nil {
// 			t.Fatal("The test has failed. Accept returned an error")
// 		}

// 		go func(c net.Conn) {
// 			defer c.Close()

// 			// Code that would handle the connection here
// 		}(conn)
// 	}
// }
