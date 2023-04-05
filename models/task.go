package models

// Task .
type Task struct {
	URL         string
	BaseURL     string
	Title       string
	Author      string
	Description string
	Image       string
	HasCover    bool
	Chapters    []*Chapter
	Rule        Rule
}

// Rule .
type Rule struct {
	RetryTimes           int
	StoryTitleDom        string
	StoryAuthorDom       string
	StoryImageDom        string
	StoryDescriptionDom  string
	ChapterListDom       string
	ChapterListStartFrom string
	ChapterDetailDom     string
}

// Chapter .
type Chapter struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Hash  string `json:"hash"`
	Path  string `json:"path"`
}
