package telegram

const msgHelp = `I can help you get your Instagram feed updates.
For logging in to your account please send your Login and Password in following format "LOG:yourloginPASS:yourpassword" 
without any spaces and "".
Your password doesn't save in any format.
After logging in use /upd to get updates.
Use /help to read this message again.
`

const msgHello = "Hi! 🤖 \n" + msgHelp

const (
	msgUnknownCommand    = "Incorrect login and password input format or Unknown command 😒"
	msgLoggedIn          = "Logged in 👌"
	msgNoNewPost         = "There are no New Post 🤷"
	msgLogInFailed       = "Log In to your account failed. 😭 Please try again."
	msgSavingAccFailed   = "Saving your account failed. 😓 Please try again"
	msgOpenAccFailed     = "Opening account failed. 🥶 Please log in again and restart"
	msgNotLoggedInBefore = "Please log in to Instagram account to get updates read /help \U0001FAF5"
)
