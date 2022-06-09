package handlersv1beta1

type Option func(*APIServer)

func WithStatsD(st StatsDClient) Option {
	return func(s *APIServer) {
		s.statsDReporter = st
	}
}
