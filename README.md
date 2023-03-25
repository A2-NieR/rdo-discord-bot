# Red Dead Online Discord Bot

A Discord bot using the [DiscordGo](https://github.com/bwmarrin/discordgo) library written for a Discord server for the game Red Dead Online.<br/>
Error monitoring via [Airbrake/Gobrake](https://github.com/airbrake/gobrake)

![image](https://user-images.githubusercontent.com/36411819/227712741-1a869ec2-a3a1-49b1-88f9-fdec4f284499.png)

This specific bot adds the following to the server:
- A message to a specific `#roles` channel to which users can react for server role self assignment via reaction emojis
- Reads the changelog.md of the repository and converts it to a message in a `#bulletin` channel
- Scans the server for a `#pc`, `#ps4` and `#xbox-one` channel, saving the channel ids for further functionality and implements 5 specific slash commands

![image](https://user-images.githubusercontent.com/36411819/227712404-93b6cc03-2513-4abf-b7bc-ff8d7de38180.png)

of which `/setup` is the first command to be used by new users (using other commands without doing the profile setup first will display a message suggesting the user to do the setup first 🤠).

![image](https://user-images.githubusercontent.com/36411819/227711884-cde0a993-e1e0-476e-bb27-b028047b0c7a.png)

After submitting, the info will get saved into your MongoDB Atlas and be updated depending on the commands the users submit:

![image](https://user-images.githubusercontent.com/36411819/227712181-f87e9d2a-8e25-48dd-9c19-e56d64c41554.png)

From then on, players can use the other commands in the channel of their platform either to flag themselves as online/offline or see if anyone else is online. When using `/online` and `/me` the bot also provides buttons for quickly updating the player's info:

![image](https://user-images.githubusercontent.com/36411819/227710657-bd5a3b31-42fb-4676-81dd-46d422ccc040.png)
