package util

// FirstOrDefault returns the first value of vargs, or the default otherwise.
func FirstOrDefault(vargs []string, def string) string {
	var path string
	if len(vargs) > 0 {
		path = vargs[0]
	} else {
		path = def
	}
	return path
}
