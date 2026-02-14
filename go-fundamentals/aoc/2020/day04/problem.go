package aoc2020day04

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

var inputFile = flag.String("inputFile", "input,txt", "Relative path to the input file")

type Passport struct {
	fields map[string]string
}

func NewPassport(rowData string) *Passport {
	passportLines := strings.Split(rowData, "\n")

	passportValues := make(map[string]string)

	for _, passportLine := range passportLines {
		parts := strings.SplitSeq(passportLine, " ")
		for part := range parts {
			fields := strings.Split(part, ":")
			passportValues[fields[0]] = fields[1]
		}
	}

	return &Passport{
		fields: passportValues,
	}
}

func (p *Passport) isValid() bool {
	requiredFields := []string{"byr", "iyr", "eyr", "hgt", "hcl", "ecl", "pid"}

	for _, requiredField := range requiredFields {
		if _, ok := p.fields[requiredField]; !ok {
			return false
		}
	}
	return true
}

func (p *Passport) isYearValid(field string, minYear, maxYear int) bool {
	v, exist := p.fields[field]
	if !exist || len(v) != 4 {
		return false
	}

	year, err := strconv.Atoi(v)
	if err != nil {
		return false
	}

	return year >= minYear && year <= maxYear
}

func (p *Passport) isBirthYearValid() bool {
	return p.isYearValid("byr", 1920, 2002)
}

func (p *Passport) isYearIssueValid() bool {
	return p.isYearValid("iyr", 2010, 2020)
}

func (p *Passport) isValidExpirationYear() bool {
	return p.isYearValid("eyr", 2020, 2030)
}

func (p *Passport) isValidHeight() bool {
	v, exist := p.fields["hgt"]
	if !exist {
		return false
	}
	var size int
	var measure string

	n, _ := fmt.Sscanf(v, "%d%s", &size, &measure)

	if n != 2 {
		return false
	}

	if measure == "cm" {
		return size >= 150 && size <= 193
	}

	if measure == "in" {
		return size >= 59 && size <= 76
	}

	return false
}

func (p *Passport) isValidHairColor() bool {
	v, exist := p.fields["hcl"]
	if !exist {
		return false
	}

	if len(v) != 7 || v[0] != '#' {
		return false
	}

	for _, c := range v[1:] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}

	return true
}

func (p *Passport) hasValidEyes() bool {
	v, exist := p.fields["ecl"]
	if !exist {
		return false
	}

	validColors := []string{"amb", "blu", "brn", "gry", "grn", "hzl", "oth"}

	if !slices.Contains(validColors, v) {
		return false
	}

	return true
}

func (p *Passport) hasValidPassportId() bool {
	v, exist := p.fields["pid"]
	if !exist {
		return false
	}

	if len(v) != 9 {
		return false
	}

	for _, c := range v {
		if !(c >= '0' && c <= '9') {
			return false
		}
	}

	return true
}

func Run() {
	flag.Parse()

	bytes, err := os.ReadFile(*inputFile)
	if err != nil {
		panic("Unable to read input file")
	}

	input := string(bytes)

	fmt.Printf("AoC2022, Day04, Part1 solution is: %d\n", part1(&input))
	fmt.Printf("AoC2022, Day04, Part2 solution is: %d\n", part2(&input))
}

func part1(input *string) int {
	lines := strings.SplitSeq(*input, "\n\n")

	total := 0
	for line := range lines {
		passport := NewPassport(line)
		if passport.isValid() {
			total += 1
		}
	}

	return total
}

func part2(input *string) int {
	lines := strings.SplitSeq(*input, "\n\n")

	total := 0
	for line := range lines {
		passport := NewPassport(line)
		if passport.hasValidEyes() && passport.hasValidPassportId() && passport.isBirthYearValid() && passport.isValidExpirationYear() && passport.isValidHairColor() && passport.isValidHeight() && passport.isYearIssueValid() {
			total += 1
		}
	}

	return total
}
