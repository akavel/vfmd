package span

type Detector interface {
	Detect(*Splitter) (consumed int)
}
