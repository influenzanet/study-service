package types

type DBConfig struct {
	URI             string
	DBNamePrefix    string
	Timeout         int
	MaxPoolSize     uint64
	IdleConnTimeout int
}

type StudyConfig struct {
	GlobalSecret               string
	TimerEventFrequency        int64 // how often the timer event should be performed (only from one instance of the service) - seconds
	TimerEventCheckIntervalMin int   // approx. how often this serice should check if to perform the timer event - seconds
	TimerEventCheckIntervalVar int   // range of the uniform random distribution - varying the check interval to avoid a steady collisions
}
