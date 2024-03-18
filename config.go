package mailbus

// Config represents the main config
type Config struct {
	DB struct {
		Type string // "bolt", "sqlite", etc.
		Path string
	}

	HTTP struct {
		Addr string
	}

	SMTP struct {
		Host     string
		Port     int
		Username string
		Password string
	}

	Newsletter struct {
		From      string
		Frequency int
		Cron      struct {
			Spec string
		}
		Product struct {
			Name string
		}
		HMAC struct {
			Secret string
		}
	}

	Sentry struct {
		DSN string
	}

	AMQP struct {
		URL string
	}
}
