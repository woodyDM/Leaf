package leaf

type Page struct {
	PageNum  int
	PageSize int
	Total    int
}

func (p Page) Offset() int {
	return (p.PageNum - 1) * p.PageSize
}
