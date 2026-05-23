package normalization

// Normalizer provides a pipeline entry point for all normalization operations.
type Normalizer struct{}

func New() *Normalizer { return &Normalizer{} }

func (n *Normalizer) Analyze(s string) []Detection { return Analyze(s) }
func (n *Normalizer) Repair(s string) string        { return RepairString(s) }
