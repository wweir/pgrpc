package pgrpc

type logger struct {
	log func(string, ...interface{})
}

// WithLogFunc set logger for pgrpc client
func WithLogFunc(f func(format string, a ...interface{})) *logger {
	return &logger{log: f}
}
func (o *logger) Log(format string, a ...interface{}) {
	if o.log != nil {
		o.log(format, a...)
	}
}
func (o *logger) applyClient(co *clientOpts) {
	co.log = o.log
}
func (o *logger) applyServer(so *serverOpts) {
	so.log = o.log
}
