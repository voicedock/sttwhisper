package config

// LanguagePack configuration
type LanguagePack struct {
	Name        string   `json:"name"`
	Languages   []string `json:"languages"`
	DownloadUrl string   `json:"download_url"`
	License     string   `json:"license"`
}
