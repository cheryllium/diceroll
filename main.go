package main

import (
  "fmt"
  "github.com/joho/godotenv"
)

func main() {
  // Load the .env file
  err := godotenv.Load(".env")
  if err != nil {
    fmt.Println("Error loading .env file")
  }

  // Initialize the DB
  InitDB()
  
  // Run the bot
  RunBot()
}
