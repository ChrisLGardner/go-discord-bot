package main

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"time"

	"github.com/honeycombio/beeline-go"
)

func rollDiceHelp() string {
	help := `Rolls dice for Chronicles of Darkness and returns number of successes.
		Format expected is:

		<number of dice>|<c> <8a|9a>

		The first number is how many dice to roll. If a chance die is needed then using "c" 
		instead of a number will roll that. For 8 again or 9 again add 8a or 9a to the end.

		Examples:

		!roll 5

		Rolls 5 dice and returns successes.

		!roll c

		Rolls a chance die.

		!roll 4 8a

		Rolls 4 dice with 8 again.
		`

	return help
}

func rollDice(ctx context.Context, dice string) (string, error) {
	ctx, span := beeline.StartSpan(ctx, "rollDice")
	defer span.Send()

	pattern := regexp.MustCompile("(?P<dice>\\d+|c) ?(?P<again>8a|9a)?")
	names := pattern.SubexpNames()
	elements := map[string]string{}

	matchingStrings := pattern.FindAllStringSubmatch(dice, -1)
	matches := []string{}

	span.AddField("rolldice.matcheslength", len(matchingStrings))
	span.AddField("rolldice.matches", matchingStrings)

	if len(matchingStrings) == 0 {
		return "", fmt.Errorf("No number of dice specified: %v", dice)
	}

	matches = matchingStrings[0]

	for i, match := range matches {
		elements[names[i]] = match
	}

	result := ""
	if elements["dice"] == "c" {
		res := roll(1, "")
		span.AddField("rolldice.dice", res)

		if res[0] == 1 {
			result = fmt.Sprintf("Dramatic Failure (%v)", res)
		} else if res[0] == 0 {
			result = fmt.Sprintf("Success (%v)", res)
		} else {
			result = fmt.Sprintf("Failure (%v)", res)
		}
	} else {
		num, _ := strconv.Atoi(elements["dice"])

		res := roll(num, elements["again"])
		span.AddField("rolldice.dice", res)

		count := 0
		for _, v := range res {
			if v >= 8 || v == 0 {
				count++
			}
		}

		span.AddField("rolldice.count", count)

		if count == 0 {
			result = fmt.Sprintf("Failure (%v)", res)
		} else if count > 5 {
			result = fmt.Sprintf("Exceptional Success (%v)", res)
		} else {
			result = fmt.Sprintf("Success (%v)", res)
		}
	}

	return result, nil
}

func roll(num int, again string) []int {

	results := []int{}
	rand.Seed(time.Now().Unix())

	for num != 0 {
		num--
		dice := rand.Intn(9)
		results = append(results, dice)
		if again == "8a" && (dice == 8 || dice == 9 || dice == 0) {
			num++
		} else if again == "9a" && (dice == 9 || dice == 0) {
			num++
		} else if dice == 0 {
			num++
		}
	}

	return results
}
