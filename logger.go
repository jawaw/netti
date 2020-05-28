package netti

// Logger .
type Logger interface {

	//Printf .
	Printf(format string, args ...interface{})

	// Debugf .
	Debugf(template string, args ...interface{})

	// Infof .
	Infof(template string, args ...interface{})

	// Warnf .
	Warnf(template string, args ...interface{})

	// Errorf .
	Errorf(template string, args ...interface{})

	// DPanicf !!!
	DPanicf(template string, args ...interface{})

	// Panicf !!!
	Panicf(template string, args ...interface{})

	// Fatalf !!!
	Fatalf(template string, args ...interface{})
}
