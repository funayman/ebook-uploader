package web

const (
	SIGWEB = signal("web-shutdown")
)

type signal string

// Signal satisfies the os.Signal interface
func (s signal) Signal() {}

// String satisfies the os.Signal interface
func (s signal) String() string {
	return string(s)
}
