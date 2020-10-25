package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/honeycombio/beeline-go"
)

var baseURI = "https://scryfall.com/search?as=grid&order=name&q="

func mtgCommand(ctx context.Context, c string) (string, error) {

	ctx, span := beeline.StartSpan(ctx, "mtgCommand")
	defer span.Send()

	beeline.AddField(ctx, "mtg.baseuri", baseURI)

	types, superTypes, c, err := getTypes(ctx, c)
	if err != nil {
		beeline.AddField(ctx, "error", err)
		return "", err
	}

	pt, c, err := getPowerToughtness(ctx, c)
	if err != nil {
		beeline.AddField(ctx, "error", err)
		return "", err
	}

	cmc, c, err := getCmc(ctx, c)
	if err != nil {
		beeline.AddField(ctx, "error", err)
		return "", err
	}

	ci, err := findColourIdentity(ctx, c)
	if err != nil {
		beeline.AddField(ctx, "error", err)
		return "", nil
	}

	uri := baseURI + "commander%3A" + ci

	if pt != "" {
		uri = uri + "+" + pt
	}
	if cmc != "" {
		uri = uri + "+cmc" + cmc
	}

	if len(superTypes) > 0 {
		for _, superType := range superTypes {
			uri = uri + "+type%3A" + superType
		}
	}

	if len(types) == 1 {
		uri = uri + "+type%3A" + types[0]
	} else if len(types) > 1 {
		uris := []string{}
		for _, i := range types {
			uris = append(uris, uri+"+type%3A"+i)
		}

		uri = ""
		for _, i := range uris {
			uri = uri + " " + i
		}
	}

	beeline.AddField(ctx, "mtgCommand.uri", uri)
	return strings.TrimSpace(uri), nil
}

func findColourIdentity(ctx context.Context, c string) (string, error) {

	ctx, span := beeline.StartSpan(ctx, "mtg.findColourIdentity")
	defer span.Send()

	c = url.QueryEscape(c)
	uri := "https://api.scryfall.com/cards/named?fuzzy=" + c

	beeline.AddField(ctx, "mtg.findColourIdentity.uri", uri)

	resp, err := http.Get(uri)
	if err != nil {
		beeline.AddField(ctx, "error", err)
		return "", err
	}

	var result map[string][]string

	json.NewDecoder(resp.Body).Decode(&result)

	ci := ""

	for _, color := range result["color_identity"] {
		ci = ci + color
	}
	beeline.AddField(ctx, "mtg.findColourIdentity", ci)

	return ci, nil
}

func getTypes(ctx context.Context, c string) (foundTypes []string, foundSuperTypes []string, remainingCommand string, err error) {

	ctx, span := beeline.StartSpan(ctx, "mtg.getTypes")
	defer span.Send()

	validTypes := []string{"artifact", "creature", "enchantment", "instant", "land", "planeswalker", "sorcery"}

	validSuperTypes := []string{"legendary", "snow", "basic", "tribal"}

	foundTypes = []string{}

	for _, validType := range validTypes {
		if match, _ := regexp.MatchString(validType, c); match == true {
			foundTypes = append(foundTypes, validType)
			c = strings.Replace(c, validType, "", 1)
		}
	}

	beeline.AddField(ctx, "mtg.foundTypes", foundTypes)

	foundSuperTypes = []string{}

	for _, validSuperType := range validSuperTypes {
		if match, _ := regexp.MatchString(validSuperType, c); match == true {
			foundSuperTypes = append(foundSuperTypes, validSuperType)
			c = strings.Replace(c, validSuperType, "", 1)
		}
	}

	beeline.AddField(ctx, "mtg.foundSuperTypes", foundSuperTypes)

	return foundTypes, foundSuperTypes, c, nil
}

func getCmc(ctx context.Context, c string) (res string, remainingCommand string, err error) {

	ctx, span := beeline.StartSpan(ctx, "mtg.getCmc")
	defer span.Send()

	pattern := regexp.MustCompile("(?:cmc)?(?P<modifier>[<=>]{0,2})(?P<cmc>\\d+)")
	equalsRegex := regexp.MustCompile("=")
	lessThanRegex := regexp.MustCompile("<")
	names := pattern.SubexpNames()
	elements := map[string]string{}

	matchingStrings := pattern.FindAllStringSubmatch(c, -1)
	matches := []string{}

	beeline.AddField(ctx, "mtg.cmc.matcheslength", len(matchingStrings))
	beeline.AddField(ctx, "mtg.cmc.matches", matchingStrings)

	if len(matchingStrings) == 0 {
		return "", c, nil
	}

	matches = matchingStrings[0]

	for i, match := range matches {
		elements[names[i]] = match
	}

	remainingCommand = strings.TrimSpace(pattern.ReplaceAllString(c, ""))

	if elements["modifier"] == "" {
		elements["modifier"] = "="
	}

	beeline.AddField(ctx, "mtg.cmc.modifier", elements["modifier"])
	beeline.AddField(ctx, "mtg.cmc.cmc", elements["cmc"])

	elements["modifier"] = equalsRegex.ReplaceAllString(elements["modifier"], "%3D")
	elements["modifier"] = lessThanRegex.ReplaceAllString(elements["modifier"], "%3C")

	res = elements["modifier"] + elements["cmc"]

	beeline.AddField(ctx, "mtg.cmc.result", res)
	beeline.AddField(ctx, "mtg.cmc.remaining", remainingCommand)

	return res, remainingCommand, nil
}

func getPowerToughtness(ctx context.Context, c string) (res string, remainingCommand string, err error) {

	ctx, span := beeline.StartSpan(ctx, "mtg.pt")
	defer span.Send()

	pattern := regexp.MustCompile("(?P<power>[\\d\\*]+)/(?P<toughness>[\\d\\*]+)")
	names := pattern.SubexpNames()
	elements := map[string]string{}

	matchingStrings := pattern.FindAllStringSubmatch(c, -1)
	matches := []string{}

	beeline.AddField(ctx, "mtg.pt.matcheslength", len(matchingStrings))
	beeline.AddField(ctx, "mtg.pt.matches", matchingStrings)

	if len(matchingStrings) == 0 {
		return "", c, nil
	}

	matches = matchingStrings[0]

	for i, match := range matches {
		elements[names[i]] = match
	}

	remainingCommand = strings.TrimSpace(pattern.ReplaceAllString(c, ""))

	res = "pow%3D" + elements["power"] + "+tou%3D" + elements["toughness"]
	beeline.AddField(ctx, "mtg.pt.result", res)
	beeline.AddField(ctx, "mtg.pt.remaining", remainingCommand)
	beeline.AddField(ctx, "mtg.pt.power", elements["power"])
	beeline.AddField(ctx, "mtg.pt.toughness", elements["toughness"])

	return res, remainingCommand, nil

}
