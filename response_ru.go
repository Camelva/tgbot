package main

type PhraseRU struct{}

// Commands
func (PhraseRU) CmdStart() string {
	// "Hello, #{username}.\n" +
	//	"I'm SoundCloud downloader bot.\n" +
	//	"Send me an url and i will respond with attached audio file"
	return "Приветствую, юзер!👋\n" +
		"Я - робот 🤖, запрограммированный скачивать музыку из SoundCloud.\n" +
		"Отправь мне ссылку и я отвечу прикрепленным аудио-файлом"
}
func (PhraseRU) CmdHelp() string {
	// "Send me an url and i will download it for you.\n" +
	//	"If something went wrong - first make sure url is valid and song available." +
	//	"Then try send message again." +
	//	"If error still persist - contact with developer (link in description)\n" +
	//	"\n===========================================\n" +
	//	"Currently supported services: \n" +
	//	"= soundcloud.com [only direct song urls yet]"
	return "Отправьте мне ссылку на песню, и я скачаю её для вас.\n" +
		"Не забывайте об ограничи" +
		"\nЕсли в процессе возникают ошибки - сперва убедитесь что ссылка рабочая, а сама песня доступная. " +
		"После этого, попробуйте отправить сообщение еще раз. " +
		"Если ошибка осталась - свяжитесь с разработчиком (ссылка в описании)\n" +
		"\n====================================\n" +
		"Список поддерживаемых сервисов:" +
		"\n====================================\n" +
		"🎵[soundcloud.com] - только прямые ссылки на песни\n" +
		"🎵[youtube.com] - только прямые ссылки на песни\n"
}
func (PhraseRU) CmdUnknown() string {
	// "I don't know that command." +
	// "Use /help for additional info"
	return "Хмм, я не знаю такой команды.\n" +
		"Посмотри в /help для дополнительной информации"
}

// Process explaining
func (PhraseRU) ProcessStart() string {
	// "Please wait..."
	return "Подожди, пытаюсь скачать эту песню 🌪"
}
func (PhraseRU) ProcessUploading() string {
	// "Everything done. Uploading song to you..."
	return "Все готово 🦾. Загружаю песню..."
}

// Exceptions
func (PhraseRU) ErrNotURL() string {
	// "Please make sure this URL is valid and try again"
	return "Эй, а это точно ссылка? 👀"
}
func (PhraseRU) ErrUndefined() string {
	// "There is some problems with this song. Please try again or contact with developer\n" +
	//	"Use /help for additional info"
	return "Хмм, с этой песней какие-то проблемы 🤔. Попробуй снова либо же свяжись с моим создателем"
}
func (PhraseRU) ErrPlaylist() string {
	// "Sorry, but i don't work with playlists yet. Use /help for more info"
	return "Это плейлист? Не люблю их... 😒"
}
func (PhraseRU) ErrUnsupportedFormat() string {
	// "This format unsupported yet. Use /help for more info"
	return "Что это за формат такой? Не похоже на песню. Иначе я бы знал что с этим делать 😧"
}
func (PhraseRU) ErrUnsupportedService() string {
	// "This service unsupported yet. Use /help for more info"
	return "Эй, я пока еще не знаком с этим сервисом 💢. Лучше посмотри в /help сначала"
}
func (PhraseRU) ErrUnavailableSong() string {
	// "Can't load this song. Make sure it is available and try again.\n" +
	//	"Otherwise, use /help for additional info"
	return "Ии... ничего. Эта песня точно доступна? Потому что я не могу её найти 😕.\n" +
		"Если ты уверен и это я ошибся - свяжись с моим создателем, может он сможет помочь.. 👀"
}
func (PhraseRU) ErrTooLarge() string {
	// "Looks like this song weighs too much.\n" +
	//		"Telegram limits uploading files size to 50mb and we can't avoid this limit.\n" +
	//		"Please try another one"
	return "Похоже что этот файл весит слишком много.\n" +
		"Телеграм ограничивает размер загружаемых файлов до 50мб и я не могу обойти это ограничение.\n" +
		"Попробуй другую песню"
}
