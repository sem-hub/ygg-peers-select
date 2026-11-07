package options

type Options struct {
	WithGit         bool
	GuessCountryYes bool
	Country         string
	DoNotRemove     bool
	TestMode        bool
	Verbose         bool
	Ipv4            bool
}

var (
	Opts Options = Options{}
)
