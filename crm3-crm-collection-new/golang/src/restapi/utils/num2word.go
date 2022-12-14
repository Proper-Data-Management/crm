// Package num2word holds Number to words converter
package utils

import (
	"fmt"
	"strings"
	"unicode"
	"strconv"
)

var ruRepl = [][]string{
	// t - тысячи; m - милионы; M - миллиарды;
	{",,,,,,", "eM"},
	{",,,,,", "em"},
	{",,,,", "et"},
	// e - единицы; d - десятки; c - сотни;
	{",,,", "e"},
	{",,", "d"},
	{",", "c"},
	{"0c0d0et", ""},
	{"0c0d0em", ""},
	{"0c0d0eM", ""},
	// --
	{"0c", ""},
	{"1c", "сто "},
	{"2c", "двести "},
	{"3c", "триста "},
	{"4c", "четыреста "},
	{"5c", "пятьсот "},
	{"6c", "шестьсот "},
	{"7c", "семьсот "},
	{"8c", "восемьсот "},
	{"9c", "девятьсот "},
	{"1d0e", "десять "},
	{"1d1e", "одиннадцать "},
	{"1d2e", "двенадцать "},
	{"1d3e", "тринадцать "},
	{"1d4e", "четырнадцать "},
	{"1d5e", "пятнадцать "},
	{"1d6e", "шестнадцать "},
	{"1d7e", "семнадцать "},
	{"1d8e", "восемнадцать "},
	{"1d9e", "девятнадцать "},
	// --
	{"0d", ""},
	{"2d", "двадцать "},
	{"3d", "тридцать "},
	{"4d", "сорок "},
	{"5d", "пятьдесят "},
	{"6d", "шестьдесят "},
	{"7d", "семьдесят "},
	{"8d", "восемьдесят "},
	{"9d", "девяносто "},
	// --
	{"0e", ""},
	{"5e", "пять "},
	{"6e", "шесть "},
	{"7e", "семь "},
	{"8e", "восемь "},
	{"9e", "девять "},
	// --
	{"1e.", "один рубль "},
	{"2e.", "два рубля "},
	{"3e.", "три рубля "},
	{"4e.", "четыре рубля "},
	{"1et", "одна тысяча "},
	{"2et", "две тысячи "},
	{"3et", "три тысячи "},
	{"4et", "четыре тысячи "},
	{"1em", "один миллион "},
	{"2em", "два миллиона "},
	{"3em", "три миллиона "},
	{"4em", "четыре миллиона "},
	{"1eM", "один миллиард "},
	{"2eM", "два миллиарда "},
	{"3eM", "три миллиарда "},
	{"4eM", "четыре миллиарда "},
	//  блок для написания копеек без сокращения "коп"
	{"11k", "11 копеек"},
	{"12k", "12 копеек"},
	{"13k", "13 копеек"},
	{"14k", "14 копеек"},
	{"1k", "1 копейка"},
	{"2k", "2 копейки"},
	{"3k", "3 копейки"},
	{"4k", "4 копейки"},
	{"k", " копеек"},
	// --
	{".", "рублей "},
	{"t", "тысяч "},
	{"m", "миллионов "},
	{"M", "миллиардов "},
}

var mask = []string{",,,", ",,", ",", ",,,,", ",,", ",", ",,,,,", ",,", ",", ",,,,,,", ",,", ","}

var sklon = make(map[string]string)


// RuMoney - деньги прописью на русском
func RuNum2Word(number interface{}, upperFirstChar bool, valCode interface{}) string {
	if number == nil {
		number = ""
	}
	if valCode == nil {
		valCode = ""
	}


	sklon["RUBкопеек"] ="копеек"
	sklon["RUBкопейка"] ="копейка"
	sklon["RUBкопейки"] ="копейки"
	sklon["RUBрублей"] ="рублей"
	sklon["RUBрубля"] ="рубля"
	sklon["RUBрубль"] ="рубль"

	sklon["KZTкопеек"] ="тиын"
	sklon["KZTкопейка"] ="тиын"
	sklon["KZTкопейки"] ="тиын"
	sklon["KZTрублей"] ="тенге"
	sklon["KZTрубля"] ="тенге"
	sklon["KZTрубль"] ="тенге"

	n,_:= strconv.ParseFloat(number.(string),16)
	s := fmt.Sprintf("%.2f",n)
	l := len(s)

	dest := s[l-3:l] + "k"
	s = s[:l-3]
	l = len(s)
	for i := l; i > 0; i-- {
		c := string(s[i-1])
		dest = c + mask[l-i] + dest
	}

	for _, r := range ruRepl {
		dest = strings.Replace(dest, r[0], r[1], -1)

	}
	if upperFirstChar {
		a := []rune(dest)
		a[0] = unicode.ToUpper(a[0])
		dest = string(a)
	}

	dest = strings.Replace(dest,"копеек",sklon[valCode.(string)+"копеек"], -1)
	dest = strings.Replace(dest,"копейка",sklon[valCode.(string)+"копейка"], -1)
	dest = strings.Replace(dest,"копейки",sklon[valCode.(string)+"копейки"], -1)
	dest = strings.Replace(dest,"рублей",sklon[valCode.(string)+"рублей"], -1)
	dest = strings.Replace(dest,"рубля",sklon[valCode.(string)+"рубля"], -1)
	dest = strings.Replace(dest,"рубль",sklon[valCode.(string)+"рубль"], -1)

	return dest
}
