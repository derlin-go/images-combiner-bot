package main

import (
	"log"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"net/http"
	"encoding/json"
	_ "image/png"
	_ "image/jpeg"
	"image"
	"fmt"
	"errors"
)

type FileInfoResult struct {
	Ok bool `json:"ok"`
	FileInfo FileInfo `json:"result"`
}

type FileInfo struct {
	FileId string `json:"file_id"`
	FilePath string `json:"file_path"`
	FileSize int `json:"file_size"`
}

func getJson(url string, target interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func (fi *FileInfo) getImage() (image.Image, string, error){
	fmt.Println("GET FILEINFO : https://api.telegram.org/file/bot227564652:AAFc7HdSDi_OhIdISU_8JqVpeXsHzmVRTck/" + fi.FilePath)
	response, err := http.Get("https://api.telegram.org/file/bot227564652:AAFc7HdSDi_OhIdISU_8JqVpeXsHzmVRTck/" + fi.FilePath)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	return image.Decode(response.Body)
}

func getImageInfo(file_id string) (*FileInfo, error){
	fileInfos := FileInfoResult{}
	fmt.Println("GET INFO : https://api.telegram.org/bot227564652:AAFc7HdSDi_OhIdISU_8JqVpeXsHzmVRTck/getFile?file_id=" + file_id)
	err := getJson("https://api.telegram.org/bot227564652:AAFc7HdSDi_OhIdISU_8JqVpeXsHzmVRTck/getFile?file_id=" + file_id, &fileInfos)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	if !fileInfos.Ok {
		return nil, errors.New("error: status not ok")
	}

	fmt.Printf("INFO: %s", fileInfos)
	return &fileInfos.FileInfo, nil

}

func main() {
	bot, err := tgbotapi.NewBotAPI("227564652:AAFc7HdSDi_OhIdISU_8JqVpeXsHzmVRTck")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		log.Println(update)
		if update.Message == nil {
			continue
		}

		if update.Message.Photo != nil {
			photos := *update.Message.Photo;
			images := make([]image.Image, len(photos))
			for i, photoSize := range(photos){
				fileInfo, err := getImageInfo(photoSize.FileID)
				if err != nil {
					log.Fatal(err)
				}else{
					var fmt string
					images[i], fmt, err = fileInfo.getImage()
					if err != nil {
						log.Fatal(err)
					}
					log.Printf(" GOT IMAGE %s (%s)\n", (*fileInfo).FilePath, fmt)
				}
			}
		}

		log.Printf("UPDATE HOOK: [%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "hello")
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}
}