package recovery

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIsBrokenConnectionError(t *testing.T) {
	tests := []struct {
		name string
		err  interface{}
		want bool
	}{
		{name: "broken pipe", err: syscall.EPIPE, want: true},
		{name: "connection reset", err: syscall.ECONNRESET, want: true},
		{name: "wrapped broken pipe", err: fmt.Errorf("write failed: %w", syscall.EPIPE), want: true},
		{name: "network operation error", err: &net.OpError{Op: "write", Net: "tcp", Err: &os.SyscallError{Syscall: "write", Err: syscall.EPIPE}}, want: true},
		{name: "other error", err: errors.New("other error"), want: false},
		{name: "non-error panic", err: "panic", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBrokenConnectionError(tt.err); got != tt.want {
				t.Fatalf("IsBrokenConnectionError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestRecoveryDoesNotWriteAgainAfterBrokenConnection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	writer := &brokenConnectionWriter{ResponseRecorder: httptest.NewRecorder()}
	request := httptest.NewRequest(http.MethodGet, "/disconnected", nil)
	router := gin.New()
	router.Use(Recovery())
	router.GET("/disconnected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	router.ServeHTTP(writer, request)

	if writer.writeCount != 1 {
		t.Fatalf("response write count = %d, want 1", writer.writeCount)
	}
}

type brokenConnectionWriter struct {
	*httptest.ResponseRecorder
	writeCount int
}

func (w *brokenConnectionWriter) Write(data []byte) (int, error) {
	w.writeCount++
	return 0, fmt.Errorf("write tcp: %w", syscall.EPIPE)
}
