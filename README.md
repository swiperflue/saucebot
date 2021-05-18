# saucebot
An image reverse searching telegram bot written in Go

## Description
This is just a simple [Telegram](https://telegram.org/) bot I wrote as an exercise to learn [the Go programming language](https://golang.org/) and [the Telegram bot API](https://core.telegram.org/bots/api), it does image reverse searches and grabs results from Google Images for general images and from Saucenao for anime images.

You can use the bot by adding it to a group and replying to a message you want to get its source with the following commands

- `/sauce` to grab the source of the image from Google Images
- `/animesauce` to grab the source of the image from saucenao.com
- reply to a non image/sticker message to search for a user's avatar

![Screenshot from 2021-05-18 01-06-31](https://user-images.githubusercontent.com/81438111/118571395-d5259a00-b775-11eb-8ef6-2498bdae6a0a.png)


## Dependencies
- golang

## Installation
1. Get your Telegram and saucenow API KEYs

- [Telegram](https://core.telegram.org/bots#3-how-do-i-create-a-bot)
- [Saucenao](https://saucenao.com/user.php)

2. From botfather add two commands to your bot `/sauce` and `/animesauce`

3. clone this repository

```
git clone https://github.com/swiperflue/saucebot.git
```

4. cd to the directory
```
cd saucebot
```

5. open the config.json file in a text editor and add your API keys
```
{
  "TelegramToken": "<your-telegram-api-key>",
  "SaucenaoToken": "<your-saucenao-api-key>",
  ...
}
```

6. build the project
```
go build .
```

7. run the bot
```
./saucebot
```
