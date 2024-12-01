package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

const DICE_MAX = 10

type DiceType int8

const (
	Normal DiceType = iota
	Hunger
)

func (t DiceType) String() string {
	switch t {
	case Normal:
		return "Normal"
	case Hunger:
		return "Hunger"
	default:
		return "Unknown"
	}
}

type RollResult int8

const (
	Fail RollResult = iota
	Success
	CriticalSuccess
	BestialFail
)

func (r RollResult) String() string {
	switch r {
	case Fail:
		return "Fail"
	case Success:
		return "Success"
	case CriticalSuccess:
		return "Critical Success"
	case BestialFail:
		return "Bestial Fail"
	default:
		return "Unknown"
	}
}

type Dice interface {
	Roll()
	GetValue() int
	GetType() DiceType
	GetResult() RollResult
}

type NormalDice struct {
	Value int
	Type  DiceType
}

func (dice *NormalDice) Roll() {
	dice.Value = rand.Intn(DICE_MAX) + 1
}
func (dice *NormalDice) GetValue() int     { return dice.Value }
func (dice *NormalDice) GetType() DiceType { return dice.Type }
func (dice *NormalDice) GetResult() RollResult {
	if dice.Value == DICE_MAX {
		return CriticalSuccess
	}
	if dice.Value > 5 && dice.Value < DICE_MAX {
		return Success
	}
	return Fail
}

type HungerDice struct {
	Value int
	Type  DiceType
}

func (dice *HungerDice) Roll() {
	dice.Value = rand.Intn(DICE_MAX) + 1
}
func (dice *HungerDice) GetValue() int     { return dice.Value }
func (dice *HungerDice) GetType() DiceType { return dice.Type }
func (dice *HungerDice) GetResult() RollResult {
	if dice.Value == DICE_MAX {
		return CriticalSuccess
	}
	if dice.Value == 1 {
		return BestialFail
	}
	if dice.Value > 5 && dice.Value < DICE_MAX {
		return Success
	}
	return Fail
}

type Game struct {
	dices          []Dice
	difficulty     int64
	hasHungerCrit  bool
	hasHungerFail  bool
	usedReroll     bool
	usedBloodSurge bool
	successes      int64
	critSuccesses  int64
	result         string
}

func (game *Game) SetUp(totalDicesCount int64, hungerDicesCount int64, difficulty int64) {
	game.difficulty = difficulty
	game.hasHungerCrit = false
	game.hasHungerFail = false
	game.usedReroll = false
	game.usedBloodSurge = false
	game.successes = 0
	game.critSuccesses = 0
	game.result = ""
	game.dices = make([]Dice, 0, totalDicesCount)

	normalDicesCount := totalDicesCount - hungerDicesCount

	var i int64
	for i = 0; i < normalDicesCount; i++ {
		game.dices = append(game.dices, &NormalDice{Type: Normal})
	}
	for i = 0; i < hungerDicesCount; i++ {
		game.dices = append(game.dices, &HungerDice{Type: Hunger})
	}
}

func (game *Game) Roll() {
	game.hasHungerCrit = false
	game.hasHungerFail = false
	game.successes = 0
	game.critSuccesses = 0
	game.result = ""
	game.usedReroll = false
	game.usedBloodSurge = false

	for _, dice := range game.dices {
		dice.Roll()
	}
	game.compute()
}

func (game *Game) compute() {
	for _, dice := range game.dices {
		switch dice.GetResult() {
		case Success:
			game.successes++
		case CriticalSuccess:
			game.critSuccesses++
			if dice.GetType() == Hunger {
				game.hasHungerCrit = true
			}
		case BestialFail:
			game.hasHungerFail = true
		}
	}

	game.successes += game.critSuccesses*2 - game.critSuccesses&1
	game.result = game.calcResult()
}

func (game *Game) calcResult() string {
	if game.difficulty == 0 {
		return game.calcDiceResult()
	}
	return game.calcAgainstDifficulty()
}

func (game *Game) calcDiceResult() string {
	if game.successes == 0 {
		return game.calcFailure()
	}
	return game.calcSuccess()
}

func (game *Game) calcAgainstDifficulty() string {
	if game.successes >= game.difficulty {
		return game.calcSuccess()
	}
	return game.calcFailure()
}

func (game *Game) calcSuccess() string {
	if game.critSuccesses < 2 {
		return "Success"
	}

	if game.hasHungerCrit == true {
		return "Messy Critical"
	}

	return "Critical Success"
}

func (game *Game) calcFailure() string {
	if game.hasHungerFail == true {
		return "Bestial Failure"
	}

	return "Failure"
}

func (game Game) String() string {
	var buffer strings.Builder
	for i, dice := range game.dices {
		fmt.Fprintf(&buffer, "%d - %s Dice. %s (%d)\n", i, dice.GetType(), dice.GetResult(), dice.GetValue())
	}
	fmt.Fprintf(&buffer, "\nRoll Result: %s! (%d)\n", game.result, game.successes)
	return buffer.String()
}

func (game *Game) CanRerollDices(diceIndexes []int64) error {
	for _, index := range diceIndexes {
		if index > int64(len(game.dices)-1) {
			return fmt.Errorf("index %d is out of range", index)
		}
		if game.dices[index].GetType() == Hunger {
			return fmt.Errorf("index %d is hunger dice", index)
		}
	}
	return nil
}

func (game *Game) CanReroll(dicesNum int64) error {
	if dicesNum <= 0 || dicesNum > 3 {
		return fmt.Errorf("incorrect dices. should be 1 - 3, got %d", dicesNum)
	}

	var normalDices int64 = 0
	for _, dice := range game.dices {
		if dice.GetType() == Normal {
			normalDices++
		}
	}

	if normalDices < dicesNum {
		return fmt.Errorf("not enough dices to reroll: wanted %d, got %d", dicesNum, normalDices)
	}

	return nil
}

func (game *Game) RerollDices(diceIndexes []int64) error {
	if game.usedReroll {
		return fmt.Errorf("Already rerolled")
	}

	game.hasHungerCrit = false
	game.hasHungerFail = false
	game.successes = 0
	game.critSuccesses = 0
	game.result = ""
	game.usedReroll = true

	for _, index := range diceIndexes {
		game.dices[index].Roll()
	}

	game.compute()
	return nil
}

func (game *Game) Reroll(dicesNum int64) error {
	if game.usedReroll {
		return fmt.Errorf("Already rerolled")
	}

	game.hasHungerCrit = false
	game.hasHungerFail = false
	game.successes = 0
	game.critSuccesses = 0
	game.result = ""
	game.usedReroll = true

	var normalDices int64 = 0

	for index := 0; normalDices < dicesNum; index++ {
		dice := game.dices[index]
		if dice.GetType() == Normal {
			dice.Roll()
			normalDices++
		}
	}

	game.compute()
	return nil
}

func (game *Game) CanBloodSurge() error {
	if len(game.dices) == 0 {
		return fmt.Errorf("Need to roll first")
	}
	if game.usedBloodSurge {
		return fmt.Errorf("Already used blood surge")
	}

	if game.usedReroll {
		return fmt.Errorf("Can't use blood surge, reroll was used")
	}

	return nil
}

func (game *Game) BloodSurge() {
	game.hasHungerCrit = false
	game.hasHungerFail = false
	game.successes = 0
	game.critSuccesses = 0
	game.result = ""
	game.usedBloodSurge = true

	dices := make([]Dice, len(game.dices)+2)
	dices[0] = &NormalDice{Type: Normal}
	dices[0].Roll()
	dices[1] = &NormalDice{Type: Normal}
	dices[1].Roll()
	copy(dices[2:], game.dices)

	game.dices = dices
	game.compute()
}

func rcCheckGame() Game {
	rcCheckGame := Game{}
	rcCheckGame.SetUp(1, 1, 0)
	rcCheckGame.Roll()

	return rcCheckGame
}

func toInts(val []string) ([]int64, error) {
	var result []int64 = make([]int64, 0, len(val))
	for _, v := range val {
		intVal, err := strconv.ParseInt(v, 10, 0)
		if err != nil {
			continue
		}

		result = append(result, intVal)
	}

	if len(result) != len(val) {
		return nil, fmt.Errorf("Can't fully convert input")
	}

	return result, nil
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var game Game = Game{}
	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}
		if input == "q" {
			break
		}

		dicesInput := strings.Split(input, " ")
		fmt.Println("-------------ROLL START-------------")
		if dicesInput[0] == "r" {
			if dicesInput[1] == "-" {
				diceIndexes, err := toInts(dicesInput[2:])
				if err != nil {
					fmt.Println(err.Error())
					continue
				}

				if err := game.CanRerollDices(diceIndexes); err != nil {
					fmt.Println(err.Error())
					continue
				}

				if err := game.RerollDices(diceIndexes); err != nil {
					fmt.Println(err.Error())
					continue
				}
			} else {
				dicesNum, err := strconv.ParseInt(dicesInput[1], 10, 0)
				if err != nil {
					fmt.Println(err.Error())
					continue
				}

				if err := game.CanReroll(dicesNum); err != nil {
					fmt.Println(err.Error())
					continue
				}
				if err := game.Reroll(dicesNum); err != nil {
					fmt.Println(err.Error())
					continue
				}
			}

			fmt.Println(game)
		} else if dicesInput[0] == "bs" {
			if err := game.CanBloodSurge(); err != nil {
				fmt.Println(err.Error())
				continue
			}
			fmt.Println("-------------ROUSE CHECK-------------")
			fmt.Println("Rouse Check", rcCheckGame())
			fmt.Println("-------------------------------------")

			game.BloodSurge()
			fmt.Println(game)
		} else if dicesInput[0] == "rc" {
			fmt.Println(rcCheckGame())
		} else {
			if len(dicesInput) < 2 || len(dicesInput) > 3 {
				fmt.Println("Wrong input")
				continue
			}

			totalDicesCount, err := strconv.ParseInt(dicesInput[0], 10, 0)
			if err != nil {
				fmt.Println("first value is NaN")
				continue
			}
			hungerDicesCount, err := strconv.ParseInt(dicesInput[1], 10, 0)
			if err != nil {
				fmt.Println("second value is NaN")
				continue
			}

			var difficulty int64
			if len(dicesInput) == 2 {
				difficulty = 0
			} else {
				difficulty, err = strconv.ParseInt(dicesInput[2], 10, 0)
				if err != nil {
					fmt.Println("difficulty value is NaN")
					continue
				}
			}

			game.SetUp(totalDicesCount, hungerDicesCount, difficulty)
			game.Roll()
			fmt.Println(game)
		}
		fmt.Println("-------------ROLL END---------------")
	}
	fmt.Println("Exit")
}
