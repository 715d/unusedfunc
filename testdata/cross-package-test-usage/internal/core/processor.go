package core

// Processor handles core processing logic
type Processor struct {
	config Config
}

// Config holds processor configuration
type Config struct {
	MaxSize int
	Prefix  string
}

// NewProcessor creates a new processor
func NewProcessor() *Processor {
	return &Processor{
		config: Config{
			MaxSize: 1024,
			Prefix:  "default",
		},
	}
}

// Transform transforms input data - USED by service package
func (p *Processor) Transform(input string) string {
	// Used by service.Service.Process
	return p.config.Prefix + ": " + input
}

// Configure updates processor configuration - UNUSED in internal package
func (p *Processor) Configure(cfg Config) {
	// This exported method in internal package is not used
	// Should be reported as unused
	p.config = cfg
}

// GetConfig returns current configuration - UNUSED in internal package
func (p *Processor) GetConfig() Config {
	// This exported method in internal package is not used
	// Should be reported as unused
	return p.config
}

// validate validates configuration - UNUSED
func (p *Processor) validate() error {
	// Unexported and unused
	// Should be reported as unused
	if p.config.MaxSize <= 0 {
		return ErrInvalidConfig
	}
	return nil
}

// Used only in tests from another package
func (p *Processor) TestOnlyMethod() string {
	// This might be used in external tests
	return "test only"
}
