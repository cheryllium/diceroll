package main

import (
  "fmt"

  "os"
  "os/signal"

  "github.com/joho/godotenv"
  "github.com/bwmarrin/discordgo"
)

/* Sets up and runs a Discord bot to respond to slash commands for rolling dice.
 * The following commands are supported: 
 * - /roll <expression> | rolls the given expression
 * @todo - /make-macro <name> <expression> | creates a macro with the given name
 * @todo - /roll-macro <name> <arguments> | rolls the macro with the given name using given arguments
 * @todo - /list-macros | lists all macros available to the server
 * @todo - /view-macro <name> | views the macro with the given name
 * @todo - /delete-macro <name> | deletes the macro with the given name
 * @todo - /edit-macro <name> <expression> | replaces existing macro with given expression
 */
func main() {
  // Load .env file
  err := godotenv.Load(".env")
  if err != nil {
    fmt.Println("Error loading .env file")
  }
  
  // Set up the discord bot
  token := os.Getenv("DISCORD_TOKEN")
  dg, err := discordgo.New("Bot " + token)
  if err != nil {
    fmt.Println("Error creating Discord session: ", err)
    return
  }

  // Open websocket connection to Discord and begin listening
  err = dg.Open()
  if err != nil {
    fmt.Println("Error opening connection: ", err)
    return
  }

  // Set up commands
  commands := []*discordgo.ApplicationCommand{
    {
      Name: "roll",
      Description: "Roll some dice",
      Options: []*discordgo.ApplicationCommandOption{
        {
          Type: discordgo.ApplicationCommandOptionString,
          Name: "expression",
          Description: "Your expression with dice notation",
          Required: true,
        },
      },
    },
  }
  commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
    "roll": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
      argument := i.ApplicationCommandData().Options[0].StringValue()
      result := ParseExpression(argument)
      
      s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
          Content: fmt.Sprintf("You asked me to roll: %s\nYou rolled a **%d**!", argument, result),
        },
      })
    },
  }

  // Register commands
  registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
  for i, v := range commands {
    cmd, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", v)
    if err != nil {
      fmt.Printf("Cannot create %v command", v.Name, err)
      return
    }
    registeredCommands[i] = cmd
  }

  // Add command handlers
  dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
    if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
      h(s, i)
    }
  })

  // Keep running this program until interrupt signal is received
  defer dg.Close()
  fmt.Println("Bot is now running. Press Ctrl+C to exit.")
  stop := make(chan os.Signal, 1)
  signal.Notify(stop, os.Interrupt)
  <- stop

  // Remove commands when exiting the program
  fmt.Println("Removing commands")
  for _, v := range registeredCommands {
    err := dg.ApplicationCommandDelete(dg.State.User.ID, "", v.ID)
    if err != nil {
      fmt.Printf("Cannot delete %v command: %v", v.Name, err)
    }
  }
}
