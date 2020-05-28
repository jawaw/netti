package configs

import (
	"testing"
	"time"
)

// TestGetDuration .
func TestGetDuration(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{"test idle_timeout", args{"redis.idle_timeout"}, 3600},
		{"test expire_time", args{"redis.expire_time"}, 1800},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDuration(tt.args.key); got != tt.want {
				t.Errorf("GetDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetInteger .
func TestGetInteger(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"test max_idle", args{"redis.max_idle"}, 1000},
		{"test max_active", args{"redis.max_active"}, 10000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetInteger(tt.args.key); got != tt.want {
				t.Errorf("GetInteger() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetString .
func TestGetString(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"test listen_addr", args{"listen_addr"}, "0.0.0.0:90"},
		{"test url", args{"redis.url"}, "10.81.9.89:16000"},
		{"test pwd", args{"redis.pwd"}, "minigame"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetString(tt.args.key); got != tt.want {
				t.Errorf("GetString() = %v, want %v", got, tt.want)
			}
		})
	}
}
