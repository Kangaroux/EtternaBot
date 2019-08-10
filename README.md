This is a discord bot for integrating with [EtternaOnline](https://etternaonline.com).

This bot is a work in progress. The API it uses is pretty horrid and in some cases the bot has to scrape web pages to get the information it needs. This has a negative impact on performance as well as readability in several areas of the code. A postgres database is used to cache most of the information the bot pulls down from the API to address some of these performance issues.

# Running the Bot

This bot runs using Docker, so if you have Docker installed then there is no setup necessary.

However, you will need a `.env` file in the root of the project that Docker will read from. The `.env` file needs two vars:
- `ETTERNA_API_KEY`: A key for the V1 EtternaOnline API.
- `BOT_TOKEN`: A discord bot token.

The project has some minimal automated testing which can be verified with `make test`.

To run the bot, first build using `make build`, then run with `make`. If the bot doesn't appear to be working, do the following:
- Run `make stop`
- Run `make db` to start the database
- Run `docker-compose up bot` to run the bot in non-daemon mode so you can easily view the logs

If the bot is running correctly you should see the message: `Bot is now running. Press CTRL-C to exit.`.

# Features

This list is subject to change (and I'll probably forget to update it) but at the moment the bot is capable of:
- Getting and comparing user profiles (MSD, ranks)
- Getting a player's most recent score
- Tracking and automatically posting recent plays which yielded a gain in rating or that had high accuracy (>99% default)
- Looking up a player's best score on the last posted song in the server (for example, bot posts a recent score by player A, player B says `;compare`, and the bot posts player B's best score for that song)