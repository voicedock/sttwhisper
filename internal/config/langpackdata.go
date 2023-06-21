package config

type LangPackData struct {
	LangPack  *LanguagePack
	ModelPath string
}

func (l *LangPackData) Downloaded() bool {
	return l.ModelPath != ""
}

func (l *LangPackData) Downloadable() bool {
	return l.LangPack.DownloadUrl != ""
}
