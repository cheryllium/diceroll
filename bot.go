package main

import (
  "fmt"
  "strings"
  
  "os"
  "os/signal"

  "github.com/bwmarrin/discordgo"
)

/* Sets up and runs a Discord bot to respond to slash commands for rolling dice.
 * The following commands are supported: 
 * - /roll <expression> | rolls the given expression
 * - /make-macro <name> <expression> | creates a macro with the given name
 * - /roll-macro <name> <arguments> | rolls the macro with the given name using given arguments
 * - /list-macros | lists all macros available to the server
 * - /view-macro <name> | views the macro with the given name
 * - /delete-macro <name> | deletes the macro with the given name
 * - /edit-macro <name> <expression> | replaces existing macro with given expression
 */
func RunBot() {
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
  }
  commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
    "roll": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
      argument := i.ApplicationCommandData().Options[0].StringValue()
      result, error := ParseExpression(argument)

      if error != nil {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
          Type: discordgo.InteractionResponseChannelMessageWithSource,
          Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("**Error**: %s", error),
          },
        })
      } else {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
          Type: discordgo.InteractionResponseChannelMessageWithSource,
          Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("You asked me to roll: %s\nYou rolled a **%d**!", argument, result),
          },
        })
      }
    },
    "make-macro": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
      name := i.ApplicationCommandData().Options[0].StringValue()
      expression := i.ApplicationCommandData().Options[1].StringValue()

      // Check if a macro with this name already exists
      existing, _ := FindMacro(i.Interaction.GuildID, name)
      if existing != nil {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
          Type: discordgo.InteractionResponseChannelMessageWithSource,
          Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("A macro with the name '%s' already exists.", name),
          },
        })
        return
      }

      // Validate the macro expression
      err := ValidateMacro(expression)
      if err != nil {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
          Type: discordgo.InteractionResponseChannelMessageWithSource,
          Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("Invalid macro expression: %s", err),
          },
        })
        return
      }

      // Create the new macro
      newMacro := Macro{
        Guild: i.Interaction.GuildID,
        Name: name,
        Expression: expression, 
      }

      MakeMacro(&newMacro)
      
      s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
          Content: fmt.Sprintf("Macro '%s' created!\nMacro expression: %s", name, expression),
        },
      })
    },
    "roll-macro": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
      name := i.ApplicationCommandData().Options[0].StringValue()
      arguments := strings.Fields(i.ApplicationCommandData().Options[1].StringValue())

      macro, _ := FindMacro(i.Interaction.GuildID, name)
      if macro != nil {
        result, err := ParseMacro(macro.Expression, arguments)

        if err != nil {
          s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
              Content: fmt.Sprintf("Error running macro: %s", err),
            },
          })
          return
        }

        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
          Type: discordgo.InteractionResponseChannelMessageWithSource,
          Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("You asked me to roll the '%s' macro.\nYou rolled a **%d**!", name, result),
          },
        })
      } else {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
          Type: discordgo.InteractionResponseChannelMessageWithSource,
          Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("No macro with the name '%s' was found.", name),
          },
        })
      }
    },
    "list-macros": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
      macros, _ := ListMacros(i.Interaction.GuildID)
      if macros != nil && len(macros) > 0 {
        listMessage := "Macros found: \n"
        for _, m := range macros {
          listMessage += fmt.Sprintf("**%s**: %s\n", m.Name, m.Expression)
        }
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
          Type: discordgo.InteractionResponseChannelMessageWithSource,
          Data: &discordgo.InteractionResponseData{
            Content: listMessage,
          },
        })
      } else {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
          Type: discordgo.InteractionResponseChannelMessageWithSource,
          Data: &discordgo.InteractionResponseData{
            Content: "No macros found. Create some with the /make-macro command.",
          },
        })
      }
    },
    "view-macro": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
      name := i.ApplicationCommandData().Options[0].StringValue()

      macro, _ := FindMacro(i.Interaction.GuildID, name)
      if macro != nil {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
          Type: discordgo.InteractionResponseChannelMessageWithSource,
          Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("Macro '%s' found: %s", macro.Name, macro.Expression),
          },
        })
      } else {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
          Type: discordgo.InteractionResponseChannelMessageWithSource,
          Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("No macro with the name '%s' was found.", name),
          },
        })
      }
    },
    "delete-macro": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
      name := i.ApplicationCommandData().Options[0].StringValue()

      macro, _ := FindMacro(i.Interaction.GuildID, name)
      if macro != nil {
        DeleteMacro(macro)
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
          Type: discordgo.InteractionResponseChannelMessageWithSource,
          Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("Macro '%s' was deleted.", name),
          },
        })
      } else {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
          Type: discordgo.InteractionResponseChannelMessageWithSource,
          Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("No macro with the name '%s' was found.", name),
          },
        })
      }
    },
    "edit-macro": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
      name := i.ApplicationCommandData().Options[0].StringValue()
      expression := i.ApplicationCommandData().Options[1].StringValue()

      macro, _ := FindMacro(i.Interaction.GuildID, name)
      if macro != nil {
        macro.Expression = expression
        EditMacro(macro)
        
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
          Type: discordgo.InteractionResponseChannelMessageWithSource,
          Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("Macro '%s' was updated: %s", name, expression),
          },
        })
      } else {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
          Type: discordgo.InteractionResponseChannelMessageWithSource,
          Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("No macro with the name '%s' was found.", name),
          },
        })
      }
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
  fmt.Println("Bot is now running. Press Ctrl+C to exit.")
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
