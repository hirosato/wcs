package model

type SupportedLang int

const (
	JA = iota
	EN = iota
)

func (lang SupportedLang) IsSupportedLang() bool {
	return lang == JA || lang == EN
}

func (lang SupportedLang) String() string {
	switch lang {
	case JA:
		return "ja"
	case EN:
		return "en"
	default:
		return "Unknown"
	}
}
