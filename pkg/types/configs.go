package types

type DBConfig struct {
	URI             string
	DBNamePrefix    string
	Timeout         int
	MaxPoolSize     uint64
	IdleConnTimeout int
}

type StudyConfig struct {
	GlobalSecret               string // GlobalSecret is used, with studySecret to create the participantID from user profileID,
	TimerEventFrequency        int64  // how often the timer event should be performed (only from one instance of the service) - seconds
	TimerEventCheckIntervalMin int    // approx. how often this serice should check if to perform the timer event - seconds
	TimerEventCheckIntervalVar int    // range of the uniform random distribution - varying the check interval to avoid a steady collisions
}

type ExternalService struct {
	Name            string           `yaml:"name"`
	URL             string           `yaml:"url"`
	APIKey          string           `yaml:"apiKey"`
	Timeout         int              `yaml:"timeout"`
	MutualTLSConfig *MutualTLSConfig `yaml:"mTLSConfig"`
}

type MutualTLSConfig struct {
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
	CAFile   string `yaml:"caFile"`
}

type ExternalServices struct {
	Services []ExternalService `yaml:"services"`
}
