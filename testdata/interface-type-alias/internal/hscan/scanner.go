package hscan

// Scanner interface defined in internal package
type Scanner interface {
	ScanRedis(s string) error
}
