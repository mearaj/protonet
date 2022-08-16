package view

func Listen[t interface{}](ch <-chan t) (t, bool) {
	shouldBreak, found := false, false
	var val t
	for {
		select {
		case val = <-ch:
			found = true
		default:
			shouldBreak = true
		}
		if shouldBreak {
			break
		}
	}
	return val, found
}
