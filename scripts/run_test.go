package scripts

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"permission/pkg/golib/v2/env"
)

func TestStart(t *testing.T) {
	const socketPath = "/usr/local/var/run/go.sock"

	c, err := net.Dial("unix", socketPath)
	assert.NoError(t, err)
	_, _ = fmt.Fprint(c, "GET /api/course/getCourseInfo?courseID=7654321 HTTP/1.0\r\n\r\n")
	scanner := bufio.NewScanner(c)
	var response string
	for scanner.Scan() {
		response += scanner.Text()
	}
	assert.Contains(t, response, "HTTP/1.0 200", "should get a 200")
	assert.Contains(t, response, "math", "resp body should match")

	t.Logf("\nresponse is : %+v\n", response)
}

func TestMain(m *testing.M) {
	env.SetRootPath("../")
	m.Run()
	os.Exit(0)
}
