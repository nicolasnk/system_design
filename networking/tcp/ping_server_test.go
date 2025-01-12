package tcp

import (
	"context"
	"io"
	"net"
	"testing"
	"time"
)

func TestPingAdvanceDeadline(t *testing.T) {
	done := make(chan struct{})
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	begin := time.Now()
	go func() {
		defer func() { close(done) }()
		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			cancel()
			conn.Close()
		}()

		resetTimer := make(chan time.Duration, 1)
		resetTimer <- time.Second
		go Pinger(ctx, conn, resetTimer)

		start_time := time.Now()
		err = conn.SetDeadline(time.Now().Add(5 * time.Second)) // 1
		if err != nil {
			t.Error(err)
			return
		}

		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				t.Logf("error at %s when reading %s", time.Since(begin).Truncate(time.Second), err)
				t.Logf("Time elapsed since last start_time set %s", time.Since(start_time).Truncate(time.Second))
				return
			}
			t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])

			resetTimer <- 0 // 2
			start_time = time.Now()
			err = conn.SetDeadline(time.Now().Add(5 * time.Second)) // 3
			if err != nil {
				t.Error(err)
				return
			}
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for i := 0; i < 4; i++ { // 4
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])
	}

	_, err = conn.Write([]byte("PONG!!!")) // should reset the ping timer // 5
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 4; i++ { // read up to four more pings // 6
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				t.Fatal(err)
			}
			break
		}
		t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])
	}
	<-done
	end := time.Since(begin).Truncate(time.Second)
	t.Logf("[%s] done", end)
	if end != 9*time.Second { // 7
		t.Fatalf("expected EOF at 9 second; actual %s", end)
	}
}
