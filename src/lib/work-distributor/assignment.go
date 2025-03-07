package workdistributor

type Assignment struct {
	Id			string	`json:"id"`
	Start		int		`json:"start"`
	End			int		`json:"end"`
	Progress	int		`json:"progress"`
	HasFailed	bool	`json:"has_failed"`
}