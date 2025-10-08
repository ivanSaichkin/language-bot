package constants

type UserState string

const (
	StateDefault          UserState = ""
	StateAwaitingWord     UserState = "awaiting_word"
	StateInReview         UserState = "in_review"
	StateInTest           UserState = "in_test"
	StateAwaitingLanguage UserState = "awaiting_language"
)

const (
	LanguageEnglish = "en"
	LanguageGerman  = "de"
	LanguageFranch  = "fr"
	LanguageSpanish = "es"
	LanguageRussian = "ru"
)

const (
	PartOfSpeechNoun      = "noun"
	PartOfSpeechVerb      = "verb"
	PartOfSpeechAdjective = "adjective"
	PartOfSpeechAdverb    = "adverb"
	PartOfSpeechPhrase    = "phrase"
)
