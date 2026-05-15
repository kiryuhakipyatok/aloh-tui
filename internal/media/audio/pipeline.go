package audio

type AudioNode interface {
	Process(buffer []int16)
}

type DSPPipeline struct {
	nodes []AudioNode
}

func NewDSPPipeline(nodesAmount int) *DSPPipeline {
	return &DSPPipeline{
		nodes: make([]AudioNode, 0, nodesAmount),
	}
}

func (dp *DSPPipeline) Process(buffer []int16) {
	for i := range dp.nodes {
		dp.nodes[i].Process(buffer)
	}
}

func (dp *DSPPipeline) AddAudioNode(node AudioNode) {
	if node != nil {
		dp.nodes = append(dp.nodes, node)
	}
}