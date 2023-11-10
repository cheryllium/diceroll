package main

import (
  "testing"
)

/* Test that a basic roll (XdY) works as expected. */
func TestParseBasicRoll(t *testing.T) {
  result, rolls, error := ParseExpression("2d20")
  if error != nil {
    t.Fatalf("Parsing 2d20 failed with error: %s", error)
  }

  if len(rolls) != 1 {
    t.Fatalf("Roll 2d20: Wrong number of diceroll expressions in result")
  }

  if rolls[0].Expression != "2d20" {
    t.Fatalf("Roll 2d20: 2d20 not found in the results")
  }

  if len(rolls[0].Results) != 2 {
    t.Fatalf("Roll 2d20: rolled %d instead of 2 dice", len(rolls[0].Results))
  }

  sum := 0
  for _, v := range rolls[0].Results {
    if v < 1 || v > 20 {
      t.Fatalf("Roll 2d20: rolled a %d which is out of range", v)
    }
    sum += v
  }

  if sum != result {
    t.Fatalf("Roll 2d20: Result %d was not the sum of the rolls (%d)", result, sum)
  }
}

/* Test that arithmetic is parsed with correct order of operations. */
func TestParseArithmeticExpression(t *testing.T) {
  result, rolls, error := ParseExpression("1+(4*6)/2")

  if error != nil {
    t.Fatalf("Parsing 1+(4*6)/2 failed with error: %s", error)
  }

  if len(rolls) != 0 {
    t.Fatalf("Parse 1+(4*6)/2: rolls was not zero")
  }

  if result != 13 {
    t.Fatalf("Parse 1+(4*6)/2: got %d instead of 13", result)
  }
}

/* Test that rolling advantage works as expected */
func TestRollAdvantage(t *testing.T) {
  for i := 0; i < 10; i++ {
    result, rolls, error := ParseExpression("2d20!")

    if error != nil {
      t.Fatalf("Parsing 2d20! failed with error: %s", error)
    }

    if len(rolls) != 1 {
      t.Fatalf("Roll 2d20!: Wrong number of diceroll expressions in result")
    }

    if rolls[0].Expression != "2d20!" {
      t.Fatalf("Roll 2d20!: 2d20! not found in the results")
    }

    if len(rolls[0].Results) != 2 {
      t.Fatalf("Roll 2d20!: rolled %d instead of 2 dice", len(rolls[0].Results))
    }
    
    max := rolls[0].Results[0]
    if max < rolls[0].Results[1] {
      max = rolls[0].Results[1]
    }
    if max != result {
      t.Fatalf("Advantage chose the lower option")
    }
  }
}

/* Test that rolling disadvantage works as expected */
func TestRollDisadvantage(t *testing.T) {
  for i := 0; i < 10; i++ {
    result, rolls, error := ParseExpression("2d20?")
    
    if error != nil {
      t.Fatalf("Parsing 2d20? failed with error: %s", error)
    }

    if len(rolls) != 1 {
      t.Fatalf("Roll 2d20?: Wrong number of diceroll expressions in result")
    }

    if rolls[0].Expression != "2d20?" {
      t.Fatalf("Roll 2d20?: 2d20? not found in the results")
    }

    if len(rolls[0].Results) != 2 {
      t.Fatalf("Roll 2d20?: rolled %d instead of 2 dice", len(rolls[0].Results))
    }

    max := rolls[0].Results[0]
    if max > rolls[0].Results[1] {
      max = rolls[0].Results[1]
    }
    if max != result {
      t.Fatalf("Disadvantage chose the higher option")
    }
  }
}

/* Test that combining dice notation with arithmetic works as expected */
func TestParseDiceWithMath(t *testing.T) {
  for i := 0; i < 20; i++ {
    result, rolls, error := ParseExpression("1d20 + 5")

    if error != nil {
      t.Fatalf("Parsing 1d20 + 5 failed with error: %s", error)
    }

    if len(rolls) != 1 {
      t.Fatalf("Roll 1d20 + 5: Wrong number of diceroll expressions in result")
    }

    if rolls[0].Expression != "1d20" {
      t.Fatalf("Roll 1d20 + 5: 1d20 not found in the results")
    }

    if len(rolls[0].Results) != 1 {
      t.Fatalf("Roll 1d20 + 5: rolled %d instead of 1 die", len(rolls[0].Results))
    }

    rollResult := rolls[0].Results[0]
    if rollResult < 1 || rollResult > 20 {
      t.Fatalf("Roll 1d20 + 5: got invalid roll result %d", rollResult)
    }
    
    expected := rollResult + 5
    if result != expected {
      t.Fatalf("Result was %d instead of %d", result, expected)
    }
  }
}

/* Test that rolling shorthand works */
func TestParseShorthand(t *testing.T) {
  expression := "d20"
  result, rolls, error := ParseExpression(expression)
  
  if error != nil {
    t.Fatalf("Parsing %s failed with error: %s", expression, error)
  }

  if len(rolls) != 1 {
    t.Fatalf("Roll %s: Wrong number of diceroll expressions in result", expression)
  }
  
  if rolls[0].Expression != "d20" {
    t.Fatalf("Roll %s: d20 not found in the results", expression)
  }

  if len(rolls[0].Results) != 1 {
    t.Fatalf("Roll %s: rolled %d instead of 1 die for d20", expression, len(rolls[0].Results))
  }

  rollResult := rolls[0].Results[0]
  if rollResult < 1 || rollResult > 20 {
    t.Fatalf("Roll %s: Roll result %d out of range", expression, rollResult)
  }

  if rollResult != result {
    t.Fatalf("Roll %s: Result of expression (%d) not what expected (%d)", expression, result, rollResult)
  }
}

/* Test that rolling shorthand ("d20") notation works in an expression */
func TestParseShorthand_1(t *testing.T) {
  expression := "2d20 + d20"
  result, rolls, error := ParseExpression(expression)

  if error != nil {
    t.Fatalf("Parsing %s failed with error: %s", expression, error)
  }

  if len(rolls) != 2 {
    t.Fatalf("Roll %s: Wrong number of diceroll expressions in result", expression)
  }

  if rolls[0].Expression != "2d20" {
    t.Fatalf("Roll %s: 2d20 not found in the results", expression)
  }
  
  if rolls[1].Expression != "d20" {
    t.Fatalf("Roll %s: d20 not found in the results", expression)
  }

  if len(rolls[0].Results) != 2 {
    t.Fatalf("Roll %s: rolled %d instead of 2 dice for 2d20", expression, len(rolls[0].Results))
  }
  
  if len(rolls[1].Results) != 1 {
    t.Fatalf("Roll %s: rolled %d instead of 1 die for d20", expression, len(rolls[1].Results))
  }

  sum := 0
  for _, v := range rolls[0].Results {
    if v < 1 || v > 20 {
      t.Fatalf("Roll %s: rolled a %d which is out of range for 2d20", expression, v)
    }
    sum += v
  }

  rollResult := rolls[1].Results[0]
  if rollResult < 1 || rollResult > 20 {
    t.Fatalf("Roll %s: got invalid roll result for d20: %d", expression, rollResult)
  }
  
  expected := rollResult + sum
  if result != expected {
    t.Fatalf("Roll %s: Result was %d instead of %d", expression, result, expected)
  }
}

func TestParseShorthand_2(t *testing.T) {
  expression := "d20 + 2d20"
  result, rolls, error := ParseExpression(expression)

  if error != nil {
    t.Fatalf("Parsing %s failed with error: %s", expression, error)
  }

  if len(rolls) != 2 {
    t.Fatalf("Roll %s: Wrong number of diceroll expressions in result", expression)
  }

  if rolls[1].Expression != "2d20" {
    t.Fatalf("Roll %s: 2d20 not found in the results", expression)
  }
  
  if rolls[0].Expression != "d20" {
    t.Fatalf("Roll %s: d20 not found in the results", expression)
  }

  if len(rolls[1].Results) != 2 {
    t.Fatalf("Roll %s: rolled %d instead of 2 dice for 2d20", expression, len(rolls[0].Results))
  }
  
  if len(rolls[0].Results) != 1 {
    t.Fatalf("Roll %s: rolled %d instead of 1 die for d20", expression, len(rolls[1].Results))
  }

  sum := 0
  for _, v := range rolls[1].Results {
    if v < 1 || v > 20 {
      t.Fatalf("Roll %s: rolled a %d which is out of range for 2d20", expression, v)
    }
    sum += v
  }

  rollResult := rolls[0].Results[0]
  if rollResult < 1 || rollResult > 20 {
    t.Fatalf("Roll %s: got invalid roll result for d20: %d", expression, rollResult)
  }
  
  expected := rollResult + sum
  if result != expected {
    t.Fatalf("Roll %s: Result was %d instead of %d", expression, result, expected)
  }
}

/* Test that filling a macro works as expected */
func TestFillMacro(t *testing.T) {
  macro := "A + (B / 2)"
  inputs := []string{"2d20", "10"}
  result := FillMacro(macro, inputs)
  if result != "2d20 + (10 / 2)" {
    t.Fatalf("FillMacro failed; gave result %s", result)
  }
}
