package record

// Record corresponds to a PTY interface stdout record
type Record struct {
  // Delay from the last record.
	Delay int `yaml:"delay"`
  // Content of the record.
	Content string `yaml:"content"`
}