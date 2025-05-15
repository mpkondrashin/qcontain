//go:build darwin

package config

const InstallationFolder = `/Users/michael/go/src/qcontain/cmd/qcontain/program`

var DefaultConfig = Config{
	Monitor: MonitorConfig{
		Folder:       `/Users/michael/go/src/qcontain/cmd/qcontain/quarantine`,
		NotifyLength: 10,
	},
	Target: TargetConfig{
		Folder:     InstallationFolder + `/Quarantine`,
		Password:   `virus`,
		Encryption: 1,
	},
	Logging: LoggingConfig{
		Level:      "info",
		File:       "qcontain.log",
		MaxSize:    10,
		MaxAge:     7,
		MaxBackups: 3,
		Compress:   true,
	},
}
