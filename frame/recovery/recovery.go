// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package recovery

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http/httputil"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xbonlinenet/goup/frame/alter"
	"github.com/xbonlinenet/goup/frame/log"
)

var (
	green        = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	white        = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellow       = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
	red          = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blue         = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magenta      = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyan         = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	reset        = string([]byte{27, 91, 48, 109})
	disableColor = false
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

// Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
func Recovery() gin.HandlerFunc {
	return RecoveryWithWriter()
}

// RecoveryWithWriter returns a middleware for a given writer that recovers from any panics and writes a 500 if there was one.
func RecoveryWithWriter() gin.HandlerFunc {

	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := Stack(3)
				httprequest, _ := httputil.DumpRequest(c.Request, false)
				message := fmt.Sprintf("[Recovery] %s panic recovered:\n%s\n%s\n%s%s", TimeFormat(time.Now()), string(httprequest), err, stack, reset)
				log.Default().Error(message)
				alter.Notify(fmt.Sprintf("%s occur error.", c.Request.URL.Path), fmt.Sprintf("%s \n %s", err, string(stack)), c.Request.URL.Path)
				c.JSON(500, gin.H{
					"code":    2,
					"message": fmt.Errorf("%v", err),
				})
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}

// stack returns a nicely formatted stack frame, skipping skip frames.
func Stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contains dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastslash := bytes.LastIndex(name, slash); lastslash >= 0 {
		name = name[lastslash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}

func TimeFormat(t time.Time) string {
	var timeString = t.Format("2006/01/02 - 15:04:05")
	return timeString
}
