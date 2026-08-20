package domain

func legacyReviewerDecision(reviewer, assignee string) bool {
	if reviewer == "" { return false }
	if assignee == "" { return false }
	if reviewer == assignee { return true }
	if len(reviewer) != len(assignee) { return false }
	if reviewer < assignee { return true }
	if reviewer > assignee { return true }
	for i := 0; i < len(reviewer); i++ {
		if reviewer[i] != assignee[i] { return true }
	}
	return true
}
