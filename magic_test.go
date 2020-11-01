package main

import (
	"context"
	"testing"
)

type TestDataItem struct {
	input    string
	result   string
	hasError bool
}

type TestDataOutputMultiArray struct {
	input    string
	result   [][]string
	hasError bool
}

type TestDataOutputArray struct {
	input    string
	result   []string
	hasError bool
}

func TestParseMtgString(t *testing.T) {

	dataItems := []TestDataItem{
		{
			"sydri, galvanic genius instant cmc<=4",
			"https://scryfall.com/search?as=grid&order=name&q=commander%3ABUW+cmc%3C%3D4+type%3Ainstant",
			false,
		},
		{
			"Linvala, Keeper of Silence creature 2/2 cmc<5",
			"https://scryfall.com/search?as=grid&order=name&q=commander%3AW+pow%3D2+tou%3D2+cmc%3C5+type%3Acreature",
			false,
		},
		{
			"omnath, locus of creation creature, sorcery",
			"https://scryfall.com/search?as=grid&order=name&q=commander%3AGRUW+type%3Acreature https://scryfall.com/search?as=grid&order=name&q=commander%3AGRUW+type%3Asorcery",
			false,
		},
		{
			"anafenza, kin-tree spirit enchantment 3",
			"https://scryfall.com/search?as=grid&order=name&q=commander%3AW+cmc%3D3+type%3Aenchantment",
			false,
		},
		{
			"atraxa, praetors' voice planeswalker cmc>5",
			"https://scryfall.com/search?as=grid&order=name&q=commander%3ABGUW+cmc>5+type%3Aplaneswalker",
			false,
		},
		{
			"child of alara legendary land 2",
			"https://scryfall.com/search?as=grid&order=name&q=commander%3ABGRUW+cmc%3D2+type%3Alegendary+type%3Aland",
			false,
		},
		{
			"chromium artifact 0",
			"https://scryfall.com/search?as=grid&order=name&q=commander%3ABUW+cmc%3D0+type%3Aartifact",
			false,
		},
		{
			"eight-and-a-half-tails snow creature",
			"https://scryfall.com/search?as=grid&order=name&q=commander%3AW+type%3Asnow+type%3Acreature",
			false,
		},
		{
			"archangel avacyn tribal instant",
			"https://scryfall.com/search?as=grid&order=name&q=commander%3ARW+type%3Atribal+type%3Ainstant",
			false,
		},
		{
			"kongming \"sleeping dragon\" creature 5",
			"https://scryfall.com/search?as=grid&order=name&q=commander%3AW+cmc%3D5+type%3Acreature",
			false,
		},
		{
			"omnath land",
			"Too many cards match ambiguous name “omnath”. Add more words to refine your search.",
			true,
		},
	}

	for _, item := range dataItems {

		ctx := context.Background()

		res, err := mtgCommand(ctx, item.input)

		if item.hasError == true {
			if err.Error() != item.result {
				t.Errorf("mtgCommand with args %v: FAILED, expected %v but got %v", item.input, item.result, err)
			}
		} else {
			if res != item.result {
				t.Errorf("mtgCommand with args %v: FAILED, expected %v but got %v", item.input, item.result, res)
			}
		}
	}
}

func TestGetTypes(t *testing.T) {

	dataItems := []TestDataOutputMultiArray{
		{
			"something instant 123",
			[][]string{{"instant"}, {}},
			false,
		},
		{
			"enchantment sorcery 13",
			[][]string{{"enchantment", "sorcery"}, {}},
			false,
		},
		{
			"legendary basic land",
			[][]string{{"land"}, {"legendary", "basic"}},
			false,
		},
	}

	for _, item := range dataItems {
		ctx := context.Background()

		types, superTypes, _, _ := getTypes(ctx, item.input)

		if len(item.result[0]) != len(types) {
			t.Errorf("getType with input %v: FAILED, expected %d number of types but got %d instead", item.input, len(item.result[0]), len(types))
		}

		for _, i := range item.result[0] {

			contains := false

			for _, j := range types {
				if i == j {
					contains = true
				}
			}

			if contains != true {
				t.Errorf("getType with input %v: FAILED, expected %v types but got %v", item.input, item.result[0], types)
			}
		}

		if len(item.result[1]) != len(superTypes) {
			t.Errorf("getType with input %v: FAILED, expected %d number of super types but got %d instead", item.input, len(item.result[1]), len(superTypes))
		}

		for _, i := range item.result[1] {

			contains := false

			for _, j := range superTypes {
				if i == j {
					contains = true
				}
			}

			if contains != true {
				t.Errorf("getType with input %v: FAILED, expected %v super types but got %v", item.input, item.result[1], superTypes)
			}
		}
	}
}

func TestGetCmc(t *testing.T) {

	dataItems := []TestDataOutputArray{
		{"instant 3", []string{"%3D3", "instant"}, false},
		{"sorcery <15", []string{"%3C15", "sorcery"}, false},
		{"enchantment >=2", []string{">%3D2", "enchantment"}, false},
	}

	for _, item := range dataItems {

		ctx := context.Background()
		cmc, remain, _ := getCmc(ctx, item.input)

		if item.result[0] != cmc {
			t.Errorf("getCmc with input %v: FAILED, expected %v but got %v", item.input, item.result[0], cmc)
		}
		if item.result[1] != remain {
			t.Errorf("getCmc with input %v: FAILED, expected %v but got %v", item.input, item.result[1], remain)
		}
	}

}

func TestGetPowerToughness(t *testing.T) {

	dataItems := []TestDataOutputArray{
		{"creature 3 1/4", []string{"pow%3D1", "tou%3D4", "creature 3"}, false},
		{"sorcery <1 */2", []string{"", "tou%3D2", "sorcery <1"}, false},
		{"enchantment 3/*", []string{"pow%3D3", "", "enchantment"}, false},
	}

	for _, item := range dataItems {

		ctx := context.Background()
		pow, tou, remain, _ := getPowerToughtness(ctx, item.input)

		if item.result[0] != pow {
			t.Errorf("getPowerToughness with input %v: FAILED, expected %v but got %v", item.input, item.result[0], pow)
		}
		if item.result[1] != tou {
			t.Errorf("getPowerToughness with input %v: FAILED, expected %v but got %v", item.input, item.result[1], tou)
		}
		if item.result[2] != remain {
			t.Errorf("getPowerToughness with input %v: FAILED, expected %v but got %v", item.input, item.result[2], remain)
		}
	}
}
