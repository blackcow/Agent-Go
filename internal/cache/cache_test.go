package cache

import (
	"testing"
	"time"
)

func TestCacheSetAndGet(t *testing.T) {
	c := New[string, int](time.Millisecond * 50)
	c.Set("key", 42)
	if val, ok := c.Get("key"); !ok || val != 42 {
		t.Fatalf("expected value 42, got %v", val)
	}
	time.Sleep(60 * time.Millisecond)
	if _, ok := c.Get("key"); ok {
		t.Fatalf("expected cache item expired")
	}
}
