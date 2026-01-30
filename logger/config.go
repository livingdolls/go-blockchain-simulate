package logger

import "go.uber.org/zap/zapcore"

type LogOutput struct {
	Type     string // "stdout", "file", "stderr"
	Level    zapcore.Level
	Path     string // for file output
	MaxSize  int    // MB
	MaxDays  int
	Compress bool
}

type Config struct {
	ServiceName string
	Env         string
	Version     string
	Level       zapcore.Level

	// File rotation
	LogPath    string
	MaxSize    int // MB
	MaxBackups int
	MaxAge     int // days
	Compress   bool

	// Queue settings
	QueueSize  int
	Workers    int
	DropOnFull bool

	// Sampling
	SampleInitial    int
	SampleThereafter int

	// Outputs
	Outputs []LogOutput
}

// ProductionConfig returns enterprise production logger config
func ProductionConfig(serviceName, version string) Config {
	return Config{
		ServiceName:      serviceName,
		Env:              "production",
		Version:          version,
		Level:            zapcore.WarnLevel,
		LogPath:          "/var/log/" + serviceName + "/app.log",
		MaxSize:          100,
		MaxBackups:       10,
		MaxAge:           30,
		Compress:         true,
		QueueSize:        10000,
		Workers:          4,
		DropOnFull:       true,
		SampleInitial:    100,
		SampleThereafter: 100,
	}
}

// DevelopmentConfig returns development logger config
func DevelopmentConfig(serviceName, version string) Config {
	return Config{
		ServiceName:      serviceName,
		Env:              "development",
		Version:          version,
		Level:            zapcore.DebugLevel,
		LogPath:          "/tmp/" + serviceName + ".log",
		MaxSize:          50,
		MaxBackups:       3,
		MaxAge:           7,
		Compress:         false,
		QueueSize:        1000,
		Workers:          2,
		DropOnFull:       false,
		SampleInitial:    100,
		SampleThereafter: 100,
	}
}
