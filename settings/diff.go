package settings

// Diff returns the difference between two string maps
func diff(old, new map[string]string) map[string]string {
	var res = make(map[string]string)
	// checks old keys
	for oldkey, oldval := range old {
		newval, ok := new[oldkey]
		if ok {
			if oldval == newval {
				continue
			}
		}
		res[oldkey] = newval
	}
	// checks for new keys
	for newkey, newval := range new {
		if _, ok := old[newkey]; !ok {
			res[newkey] = newval
		}
	}
	return res
}
