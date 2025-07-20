package models

type Pagination struct {
	Limit  int   `json:"limit" form:"limit"`
	Page   int   `json:"page" form:"page"`
	Total  int64 `json:"total"`
	Offset int   `json:"offset"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}
