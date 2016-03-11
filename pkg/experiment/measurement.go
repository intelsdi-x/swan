package experiment

type Measurement struct {
	// Service Level Indicators for 5% (0 index), 10%, ... 95% (18 index) load Points.
	// Currently they are in form of the 99p latencies in us
	// TODO(bplotka): Move that to map for clear view?
	// TODO(bplotka): Decide: Float vs Uint?
	SliLoadPointLatencies99pUs [19]float64
}
