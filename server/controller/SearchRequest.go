package controller

type SearchRequest struct {
	Title    string `json:"title" form:"title"`
	Author   string `json:"author" form:"author"`
	MinLevel int    `json:"min_level" form:"min_level"`
	MaxLevel int    `json:"max_level" form:"max_level"`
	PageNo   int    `json:"page_no" form:"page_no"`
	PageSize int    `json:"page_size" form:"page_size"`
}
