//go:build windows

package config

const InstallationFolder = `C:\Program Files\QContain`

var DefaultConfig = Config{
	Monitor: MonitorConfig{
		Folder:       `C:\Program Files\Trend\SProtect\x64\Virus`,
		NotifyLength: 10,
	},
	Target: TargetConfig{
		Folder:     InstallationFolder + `\Quarantine`,
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
