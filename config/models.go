package config

type Logger struct {
	LogFileName                string `yaml:"log_file_name"`
	LoggingLevel               string `yaml:"logging_level"`
	LogFileSizeCappingInMBs    int    `yaml:"log_file_size_capping_in_mbs"`
	MaxLogBackupsCount         int    `yaml:"max_log_backups_count"`
	MaxOldLogRetentionInDays   int    `yaml:"max_old_log_retention_in_days"`
	OldLogsCompressionRequired bool   `yaml:"logs_compression_required"`
}

type Config struct {
	Logger Logger `yaml:"logger"`
}
