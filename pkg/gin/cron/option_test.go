package cron

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestWithLocation(t *testing.T) {
	c := New(gin.New(), WithLocation(time.UTC))
	if c.location != time.UTC {
		t.Errorf("expected UTC, got %v", c.location)
	}
}

func TestWithParser(t *testing.T) {
	var parser = NewParser(Dow)
	c := New(gin.New(), WithParser(parser))
	if c.parser != parser {
		t.Error("expected provided parser")
	}
}
