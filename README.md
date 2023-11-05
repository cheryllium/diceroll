# DiceMancer Setup

## Create application on Discord

1. Ensure you have Developer Mode enabled in your Discord settings (under Advanced)
2. Go to: [https://discord.com/developers/applications](https://discord.com/developers/applications)
3. Create a new application using the "New Application" button

## Add bot to a testing server

4. On the left sidebar, click OAuth2 -> URL Generator
5. Under Scopes, select `applications.commands` and `bot` scopes
6. Under Bot Permissions, select Send Messages (in the middle column)
7. Copy the generated URL and use it to add your bot to the desired testing server

## Configure the application

1. On the left sidebar, click Bot to go to the bot's settings and click Reset Token. Copy the generated token; you will only see it once!
2. Copy the .env.example file into a file named .env and paste your token after `DISCORD_TOKEN=` in the new .env file
3. In Discord, right-click the icon for your desired testing server and select Copy Server ID
4. Copy allowed_servers.example.json into a file named allowed_servers.json, and edit it to have your server ID instead of the example server ID

After this, you should be able to run the application and use it from your Discord server. 
