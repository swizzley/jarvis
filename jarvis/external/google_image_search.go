package external

// GoogleImage is a result from google image search.
type GoogleImage struct {
}

// GoogleImageSearch searches google images.
func GoogleImageSearch(searchStr string) ([]GoogleImage, error) {
	//queryURL := fmt.Sprintf("http://ajax.googleapis.com/ajax/services/search/images?v=1.0&safe=active&as_filetype=gif&rsz=8&imgsz=medium&q=animated+%s", searchStr)
	return []GoogleImage{}, nil
}
