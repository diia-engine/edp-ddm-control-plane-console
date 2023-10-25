package locale

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var bundle = i18n.NewBundle(language.English)
var locUk *i18n.Localizer
var locEn *i18n.Localizer

func init() {
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.LoadMessageFile("locale/en.json")
	bundle.LoadMessageFile("locale/uk.json")
	locUk = i18n.NewLocalizer(bundle, "uk")
	locEn = i18n.NewLocalizer(bundle, "en")
}

func getLocalizer() *i18n.Localizer {
	if os.Getenv("LANGUAGE") == "en" {
		return locEn
	}
	return locUk
}

func Localize(word ...string) string {
	message := strings.TrimSpace(strings.Join(word, " "))
	localizer := getLocalizer()
	return localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: message,
	})
}

func LocalizeWithParams(message string, data map[string]interface{}) string {
	localizer := getLocalizer()
	return localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:    message,
		TemplateData: data,
	})
}

func DateFormat() string {
	if os.Getenv("LANGUAGE") == "en" {
		return "01/02/2006"
	}
	return "02.01.2006"
}

func DateTimeFormat() string {
	if os.Getenv("LANGUAGE") == "en" {
		return "01/02/2006 3:04 PM"
	}
	return "02.01.2006 15:04"
}

func DateTimePreciseFormat() string {
	if os.Getenv("LANGUAGE") == "en" {
		return "01/02/2006 3:04:05 PM"
	}
	return "02.01.2006 15:04:05"
}
