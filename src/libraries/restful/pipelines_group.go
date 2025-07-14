package restful

type PipelineGroups struct {
	groupNames     []string
	pipelineGroups map[string][]Pipeline
}

func NewPipelineGroups() *PipelineGroups {
	group := new(PipelineGroups)
	group.pipelineGroups = make(map[string][]Pipeline)
	return group
}

func (group *PipelineGroups) Push(groupName string, pipelines ...Pipeline) {
	if _, exists := group.pipelineGroups[groupName]; !exists {
		group.groupNames = append(group.groupNames, groupName)
	}
	group.pipelineGroups[groupName] = append(group.pipelineGroups[groupName], pipelines...)
}

func (group *PipelineGroups) ExecuteAll(req *Request, res *Response) {
	merged := []Pipeline{}
	for _, groupName := range group.groupNames {
		merged = append(merged, group.pipelineGroups[groupName]...)
	}

	for i := range merged {
		if i != len(merged)-1 {
			merged[i].SetNext(merged[i+1])
		}
	}

	if len(merged) > 0 {
		merged[0].Handle(req, res)
	}
}
