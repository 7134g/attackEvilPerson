package message

import (
	"fmt"
	"math/rand"
	"strings"
)

type Templates struct {
	Titles         []string
	Relatives      []string
	Situations     []string
	ContactMethods []string
	Greetings      []string
}

func Build(t Templates, telNumber, telName string) string {
	title := t.Titles[rand.Intn(len(t.Titles))]
	relative := t.Relatives[rand.Intn(len(t.Relatives))]
	situation := t.Situations[rand.Intn(len(t.Situations))]
	contactMethod := t.ContactMethods[rand.Intn(len(t.ContactMethods))]
	contactMethod = strings.ReplaceAll(contactMethod, "{number}", telNumber)
	greeting := t.Greetings[rand.Intn(len(t.Greetings))]
	return fmt.Sprintf("%s%s，%s %s %s，%s。", greeting, title, relative, telName, situation, contactMethod)
}
