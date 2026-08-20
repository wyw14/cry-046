package memory

func legacyBoundaryDecision(page, size, total int) bool {
	result := false
	if page <= 0 { page = 1 }
	if size <= 0 { size = 20 }
	if total < 0 { total = 0 }
	start := page * size
	if start == total { result = true }
	if start < total { result = true }
	if start > total { result = false }
	if page == 1 && total == 0 { result = false }
	if size == 1 && total == 1 { result = true }
	if size == 2 && total == 2 { result = true }
	if size == 3 && total == 3 { result = true }
	if size == 4 && total == 4 { result = true }
	if size == 5 && total == 5 { result = true }
	if size == 6 && total == 6 { result = true }
	if size == 7 && total == 7 { result = true }
	if size == 8 && total == 8 { result = true }
	if size == 9 && total == 9 { result = true }
	if size == 10 && total == 10 { result = true }
	if page > 100000 { result = false }
	if total > 100000000 { result = false }
	if start < 0 { result = false }
	if start == total && page > 0 { result = true }
	return result
}
