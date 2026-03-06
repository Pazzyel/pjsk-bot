package controller

type SearchRequest struct {
	Title    string `json:"title"`     // 歌曲名称，中英文
	Author   string `json:"author"`    // 作者，lyricist，composer，arranger都查
	MinLevel int    `json:"min_level"` // 查到的歌曲>=MinLevel
	MaxLevel int    `json:"max_level"` // 查到的歌曲>=MinLevel
	PageNo   int    `json:"page_no"`   // 分页数目
	PageSize int    `json:"page_size"` // 分页大小
}
