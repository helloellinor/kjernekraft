package config

import (
	"sync"
	"time"
)

// GlobalSettings holds application-wide configuration
type GlobalSettings struct {
	mu       sync.RWMutex
	Timezone string
	Location *time.Location
}

var instance *GlobalSettings
var once sync.Once

// GetInstance returns the singleton instance of GlobalSettings
func GetInstance() *GlobalSettings {
	once.Do(func() {
		instance = &GlobalSettings{
			Timezone: "Europe/Oslo", // Default timezone
		}
		// Load Oslo timezone by default
		loc, err := time.LoadLocation(instance.Timezone)
		if err != nil {
			// Fallback to UTC if Oslo timezone fails
			loc = time.UTC
			instance.Timezone = "UTC"
		}
		instance.Location = loc
	})
	return instance
}

// GetTimezone returns the current timezone setting
func (gs *GlobalSettings) GetTimezone() string {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.Timezone
}

// GetLocation returns the current time.Location
func (gs *GlobalSettings) GetLocation() *time.Location {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.Location
}

// SetTimezone updates the timezone setting
func (gs *GlobalSettings) SetTimezone(timezone string) error {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return err
	}
	
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.Timezone = timezone
	gs.Location = loc
	return nil
}

// GetCurrentTime returns the current time in the configured timezone
func (gs *GlobalSettings) GetCurrentTime() time.Time {
	return time.Now().In(gs.GetLocation())
}