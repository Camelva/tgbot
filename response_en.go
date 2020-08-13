package main

type PhraseEN struct{}

// Commands
func (PhraseEN) CmdStart() string {
	return "Hello, #{username}.\n" +
		"I'm SoundCloud downloader bot.\n" +
		"Send me an url and i will respond with attached audio file"
}
func (PhraseEN) CmdHelp() string {
	return "Send me an url and i will download it for you.\n" +
		"If something went wrong - first make sure url is valid and song available." +
		"Then try send message again." +
		"If error still persist - contact with developer (link in description)\n" +
		"\n===========================================\n" +
		"Currently supported services: \n" +
		"🎵[soundcloud.com] - only direct song links yet\n" +
		"🎵[youtube.com] - only direct song links yet\n"
}
func (PhraseEN) CmdUnknown() string {
	return "I don't know that command. " +
		"Use /help for additional info"
}

// Process explaining
func (PhraseEN) ProcessStart() string {
	return "Please wait..."
}
func (PhraseEN) ProcessUploading() string {
	return "Everything done. Uploading song to you..."
}

// Exceptions
func (PhraseEN) ErrNotURL() string {
	return "Please make sure this URL is valid and try again"
}
func (PhraseEN) ErrUndefined() string {
	return "There is some problems with this song. Please try again or contact with developer\n" +
		"Use /help for additional info"
}
func (PhraseEN) ErrPlaylist() string {
	return "Sorry, but i don't work with playlists yet. Use /help for more info"
}
func (PhraseEN) ErrUnsupportedFormat() string {
	return "This format unsupported yet. Use /help for more info"
}
func (PhraseEN) ErrUnsupportedService() string {
	return "This service unsupported yet. Use /help for more info"
}
func (PhraseEN) ErrUnavailableSong() string {
	return "Can't load this song. Make sure it is available and try again.\n" +
		"Otherwise, use /help for additional info"
}
func (PhraseEN) ErrTooLarge() string {
	return "Looks like this song weighs too much.\n" +
		"Telegram limits uploading files size to 50mb and we can't avoid this limit.\n" +
		"Please try another one"
}
