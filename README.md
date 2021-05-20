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
- [tucnak/telebot](https://github.com/tucnak/telebot)

## Installation
1. Get your Telegram and saucenow API KEYs

- [Telegram](https://core.telegram.org/bots#3-how-do-i-create-a-bot)
- [Saucenao](https://saucenao.com/user.php)

2. From botfather add two commands to your bot `/sauce` and `/animesauce` and disable privacy mode

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
## Usage
### Manual
Here's how you run the bot manually:
```
nohup ./saucebot & disown
nohup ./saucebot >/dev/null 2>&1 & disown # If you want to disable output
```
### Systemd Service
Alternatively, you can create a new systemd service, which handles the bot
restart in a way more neat way, with these commands:
```
sed -i 's/userplaceholder/BOTUSER/' saucebot.service
sed -i 's/pathplaceholder/BOTPATH/g' saucebot.service
sudo cp saucebot.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo service saucebot start
```
And to check the bot status you can simply type:
```
sudo service saucebot status
```
or, for reading the full log:
```
sudo journalctl -u saucebot.service
```
### Docker/Podman
If you like docker or podman, you can easily build the container using the
Dockerfile:
```
sudo docker build -m saucebot .
```
Then you can easily start the bot with:
```
sudo docker run -it saucebot
```

