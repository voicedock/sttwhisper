package config

// Model configuration
type Model struct {
	Name        string   `json:"name"`
	Languages   []string `json:"languages"`
	DownloadUrl string   `json:"download_url"`
	License     string   `json:"license"`
}

type ModelWrap struct {
	Model      *Model
	ModelPath  string
	Downloaded bool
}

func (l *ModelWrap) Downloadable() bool {
	return l.Model.DownloadUrl != ""
}
