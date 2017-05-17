package baymax


type Baymax struct {

}

type Conf struct {
	Targets []Target
	LogLines int
	LogFile string
}

type Target struct {
	Name string
	URL string
	LogLocation string
}

func NewBaymax() (*Baymax, error) {
	return &Baymax{}, nil
}
