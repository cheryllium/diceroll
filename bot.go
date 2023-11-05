package main

import (
  "fmt"
  "strings"
  
  "os"
  "os/signal"

  "github.com/bwmarrin/discordgo"
)

/* Formats the result of a dice roll in a pretty, human-readable way.
 */
func formatRollResult(expression string, result int, rolls []DiceRoll) string {
  rollResults := ""
  for _, r := range rolls {
    rollResults += fmt.Sprintf("> 🎲 **%s** %v\n", r.Expression, r.Results)
  }
  return fmt.Sprintf(
    "You asked me to roll: %s\nYou rolled a **%d**!\n> *ROLL RESULTS*\n%s",
    expression,
    result,
    rollResults,
  )
}

/* Sends a message to Discord. 
 * Used for the bot to respond to slash commands. 
 */
func sendDiscordMessage(s* discordgo.Session, i *discordgo.InteractionCreate, message string) {
  s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
    Type: discordgo.InteractionResponseChannelMessageWithSource,
    Data: &discordgo.InteractionResponseData{
      Content: message,
    },
  })
}

/* Sets up and runs a Discord bot to respond to slash commands for rolling dice.
 * The following commands are supported: 
 * - /roll <expression> | rolls the given expression
 * - /make-macro <name> <expression> | creates a macro with the given name
 * - /roll-macro <name> <arguments> | rolls the macro with the given name using given arguments
 * - /list-macros | lists all macros available to the server
 * - /view-macro <name> | views the macro with the given name
 * - /delete-macro <name> | deletes the macro with the given name
 * - /edit-macro <name> <expression> | replaces existing macro with given expression
 * - /help-me-roll | displays help/usage information
 */
func RunBot() {
  // Set up the discord bot
  fmt.Println("Initializing bot...")
  token := os.Getenv("DISCORD_TOKEN")
  dg, err := discordgo.New("Bot " + token)
  if err != nil {
    fmt.Println("Error creating Discord session: ", err)
    return
  }

  // Open websocket connection to Discord and begin listening
  fmt.Println("Opening websocket connection...")
  err = dg.Open()
  if err != nil {
    fmt.Println("Error opening connection: ", err)
    return
  }

  // Set up commands
  fmt.Println("Registering commands...")
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
    {
      Name: "make-macro",
      Description: "Create a macro",
      Options: []*discordgo.ApplicationCommandOption{
        {
          Type: discordgo.ApplicationCommandOptionString,
          Name: "name",
          Description: "The name of the macro",
          Required: true,
        },
        {
          Type: discordgo.ApplicationCommandOptionString,
          Name: "expression",
          Description: "The macro expression, using A, B, C etc for inputs to the macro",
          Required: true,
        },
      },
    },
    {
      Name: "roll-macro",
      Description: "Roll one of your custom macros",
      Options: []*discordgo.ApplicationCommandOption{
        {
          Type: discordgo.ApplicationCommandOptionString,
          Name: "name",
          Description: "The name of the macro you want to roll",
          Required: true,
        },
        {
          Type: discordgo.ApplicationCommandOptionString,
          Name: "inputs",
          Description: "The inputs to the macro, separated by spaces",
          Required: true,
        },
      },
    },
    {
      Name: "list-macros",
      Description: "List all macros available to the server",
    },
    {
      Name: "view-macro",
      Description: "View an existing macro",
      Options: []*discordgo.ApplicationCommandOption{
        {
          Type: discordgo.ApplicationCommandOptionString,
          Name: "name",
          Description: "The name of the macro",
          Required: true,
        },
      },
    },
    {
      Name: "delete-macro",
      Description: "Delete an existing macro",
      Options: []*discordgo.ApplicationCommandOption{
        {
          Type: discordgo.ApplicationCommandOptionString,
          Name: "name",
          Description: "The name of the macro",
          Required: true,
        },
      },
    },
    {
      Name: "edit-macro",
      Description: "Create a macro",
      Options: []*discordgo.ApplicationCommandOption{
        {
          Type: discordgo.ApplicationCommandOptionString,
          Name: "name",
          Description: "The name of the macro",
          Required: true,
        },
        {
          Type: discordgo.ApplicationCommandOptionString,
          Name: "expression",
          Description: "The macro expression, using A, B, C etc for inputs to the macro",
          Required: true,
        },
      },
    },
    {
      Name: "help-me-roll",
      Description: "Shows you how to use the DiceMancer bot",
    },
  }
  commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
    "roll": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
      argument := i.ApplicationCommandData().Options[0].StringValue()
      result, rolls, error := ParseExpression(argument)

      if error != nil {
        sendDiscordMessage(s, i, fmt.Sprintf("**Uh-oh!** Error occurred parsing: %s \n%s", argument, error))
      } else {
        sendDiscordMessage(s, i, formatRollResult(argument, result, rolls))
      }
    },
    "make-macro": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
      name := i.ApplicationCommandData().Options[0].StringValue()
      expression := i.ApplicationCommandData().Options[1].StringValue()

      // Check if a macro with this name already exists
      existing, _ := FindMacro(i.Interaction.GuildID, name)
      if existing != nil {
        sendDiscordMessage(s, i, fmt.Sprintf("A macro with the name '%s' already exists.", name))
        return
      }

      // Validate the macro expression
      err := ValidateMacro(expression)
      if err != nil {
        sendDiscordMessage(s, i, fmt.Sprintf("Invalid macro expression: %s", err))
        return
      }

      // Create the new macro
      newMacro := Macro{
        Guild: i.Interaction.GuildID,
        Name: name,
        Expression: expression, 
      }

      MakeMacro(&newMacro)
      sendDiscordMessage(s, i, fmt.Sprintf("Macro '%s' created!\nMacro expression: %s", name, expression))
    },
    "roll-macro": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
      name := i.ApplicationCommandData().Options[0].StringValue()
      arguments := strings.Fields(i.ApplicationCommandData().Options[1].StringValue())

      macro, _ := FindMacro(i.Interaction.GuildID, name)
      if macro != nil {
        expression := FillMacro(macro.Expression, arguments)
        result, rolls, err := ParseExpression(expression)

        if err != nil {
          sendDiscordMessage(s, i, fmt.Sprintf("**Uh-oh!** Error occurred parsing: %s \n%s", expression, err))
          return
        }
        sendDiscordMessage(s, i, formatRollResult(expression, result, rolls))
      } else {
        sendDiscordMessage(s, i, fmt.Sprintf("No macro with the name '%s' was found.", name))
      }
    },
    "list-macros": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
      macros, _ := ListMacros(i.Interaction.GuildID)
      if macros != nil && len(macros) > 0 {
        listMessage := "Macros found: \n"
        for _, m := range macros {
          listMessage += fmt.Sprintf("**%s**: %s\n", m.Name, m.Expression)
        }
        sendDiscordMessage(s, i, listMessage)
      } else {
        sendDiscordMessage(s, i, "No macros found. Create some with the /make-macro command.")
      }
    },
    "view-macro": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
      name := i.ApplicationCommandData().Options[0].StringValue()

      macro, _ := FindMacro(i.Interaction.GuildID, name)
      if macro != nil {
        sendDiscordMessage(s, i, fmt.Sprintf("Macro '%s' found: %s", macro.Name, macro.Expression))
      } else {
        sendDiscordMessage(s, i, fmt.Sprintf("No macro with the name '%s' was found.", name))
      }
    },
    "delete-macro": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
      name := i.ApplicationCommandData().Options[0].StringValue()

      macro, _ := FindMacro(i.Interaction.GuildID, name)
      if macro != nil {
        DeleteMacro(macro)

        sendDiscordMessage(s, i, fmt.Sprintf("Macro '%s' was deleted.", name))
      } else {
        sendDiscordMessage(s, i, fmt.Sprintf("No macro with the name '%s' was found.", name))
      }
    },
    "edit-macro": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
      name := i.ApplicationCommandData().Options[0].StringValue()
      expression := i.ApplicationCommandData().Options[1].StringValue()

      macro, _ := FindMacro(i.Interaction.GuildID, name)
      if macro != nil {
        macro.Expression = expression
        EditMacro(macro)
        
        sendDiscordMessage(s, i, fmt.Sprintf("Macro '%s' was updated: %s", name, expression))
      } else {
        sendDiscordMessage(s, i, fmt.Sprintf("No macro with the name '%s' was found.", name))
      }
    },
    "help-me-roll": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
      helpMessage := `**DiceMancer Bot Available Commands**

🎲 Basic Usage  🎲
**/roll** <expression>
- Example usage: `+"`"+`/roll 4d10 + 5`+"`"+`
- You can give it any arithmetic expression with both numbers and dice notation.
- Dice notation must be in the form XdY, where X and Y are integers.
- For advantage and disadvantage, you can write ! or ? after your dice notation to get the highest and lowest roll respectively. For example, 4d10! will get the highest of the four rolls, while 4d10? will get the lowest.
- You can roll up to d200 and up to 20 rolls at once.

🎲 Macros  🎲
A macro is an expression you can re-use again and again. Macros can have inputs, which must be written as uppercase letters starting from A. If the macro only has one input, it must be named A; two, must be named A and B, and so on.

For example, you can have a macro: `+"`"+`4 * (A + B)`+"`"+`
You will be able to roll this macro substituting anything you'd like for the variables A and B.

**/make-macro** <name> <expression>
- This is used to create a macro. For example: `+"`"+`/make-macro my-macro 4 * (A + B)`+"`"+`
- Macros can be named anything, with a maximum of 128 characters.

**/roll-macro** <name> <inputs separated by spaces>
- This is how you roll a macro once it's created. Specify the name of the macro, following by what you want the A, B, C, etc to be separated by spaces. (They can be either numbers or dice notation.)
- For example: `+"`"+`/roll-macro my-macro 10 4d6`+"`"+`

There are several other commands to help you view, edit, and delete macros: 
**/list-macros** | Lists all macros available.
**/view-macro** <name> | Displays the macro with the given name.
**/delete-macro** <name> | Deletes the macro with the given name.
**/edit-macro** <name> <expression> | Updates the existing macro.

Macros are tied to the server and macros created by this server can only be used in this server. 

Please enjoy using DiceMancer, and feel free to contact the developer <@284867832376721409> if you have further questions or comments.
`
      sendDiscordMessage(s, i, helpMessage)
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
  fmt.Println("Adding command handlers...")
  dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
    // Check if server is allowed to use the bot
    serverHasAccess, err := ServerHasAccess(i.Interaction.GuildID)
    if err != nil {
      s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
          Content: fmt.Sprintf("Error loading list of allowed servers. Please contact bot admin for support."),
        },
      })
      return
    }
    if !serverHasAccess {
      s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
          Content: fmt.Sprintf("Your server does not have access to use the bot."),
        },
      })
      return
    }
    
    if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
      h(s, i)
    }
  })

  // Keep running this program until interrupt signal is received
  defer dg.Close()
  fmt.Println("** Bot is now running. Press Ctrl+C to exit. **")
  stop := make(chan os.Signal, 1)
  signal.Notify(stop, os.Interrupt)
  <- stop

  // Uncomment if you'd like to remove all commands when exiting the program
  /*
  fmt.Println("Removing commands")
  for _, v := range registeredCommands {
    err := dg.ApplicationCommandDelete(dg.State.User.ID, "", v.ID)
    if err != nil {
      fmt.Printf("Cannot delete %v command: %v", v.Name, err)
    }
  }
  */
}
