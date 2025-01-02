package failover

import "time"

func OptCheckUrlBeforeAdding(check bool) func(*Options) {
	return func(options *Options) {
		options.CheckUrlBeforeAdding = check
	}
}

func OptCheckUrlDelay(delay time.Duration) func(*Options) {
	return func(options *Options) {
		options.CheckUrlDelay = delay
	}
}

func OptReqOnErr(onErr onErrType) func(*Options) {
	return func(options *Options) {
		options.ReqOnErr = onErr
	}
}

func OptMaxAttempts(maxAttempts uint16) func(*Options) {
	return func(options *Options) {
		options.MaxAttempts = maxAttempts
	}
}
