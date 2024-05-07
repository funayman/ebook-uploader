package web

type CORSOptions struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge         int
}

func NewCorsOptions(origins ...string) CORSOptions {
	return CORSOptions{
		AllowedOrigins: origins,
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "Content-Length", "X-CSRF-Token"},
		MaxAge:         86400,
	}
}
