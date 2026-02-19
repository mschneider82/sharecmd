package urlshortener

// URLShortener Interface
type URLShortener interface {
	GetName() string
	SetupQuestions() map[string]string
	ShortURL() (string, error)
}
