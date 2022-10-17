package telegram

const msgHelp = `I can help you get your Instagram feed updates.
For saving your posts use /startauth command and give permissions for app.
And repost the message received in the browser to the chat bot.
After authorization use /getposts to receive you posts (authorization token is only valid for 1 hour, then you will need to authorizate again).
If you want to get your feed updates send Login and Password in following format "LOG:yourloginPASS:yourpassword" 
without any spaces and "".
Your password doesn't save in any format.
After logging in, use /upd to get updates.
Use /help to read this message again.
`

const msgHello = "Hi! 🤖 \n" + msgHelp

const (
	msgUnknownCommand    = "Incorrect login and password input format or Unknown command 😒"
	msgLoggedIn          = "Logged in 👌"
	msgNoNewPost         = "There are no New Post 🤷"
	msgLogInFailed       = "Log In to your account failed. 😭 Please try again"
	msgSavingAccFailed   = "Saving your account failed. 😓 Please try again"
	msgOpenAccFailed     = "Opening account failed. 🥶 Please log in again and restart"
	msgNotLoggedInBefore = "Please log in to Instagram account to get updates read /help \U0001FAF5"
	msgSuccessfulAuth    = "Authentication successful, now you can save posts 😎 (/getposts) "
	msgAuthFailed        = "Authentication failed. 😭 Please try again"
	msgCantGetPosts      = "I can't get posts. 😥 Please retry or authenticate your account again. Access token works only 1 hour ⏳"
)
