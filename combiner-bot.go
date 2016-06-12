package main

import (
    "log"
    "github.com/go-telegram-bot-api/telegram-bot-api"
    "net/http"
    _ "image/png"
    _ "image/jpeg"
    "image"
    "fmt"
    "github.com/derlin-go/combiner"
)

const (
    SESSION_TIMEOUT = 18000 // ms
    MAX_IMAGES = 10
    CMD_START = "/start"
    CMD_GENERATE = "/gen"
    CMD_ABORT = "/stop"
)

var ACCEPTED_MIMETYPE = map[string]bool{"image/jpg": true, "image/jpeg": true, "image/png":true}

type BotImage struct {
    Img        image.Image
    Format     string
    DirectLink string
}

type Session struct {
    UserName string
    Images   []*image.Image
    NbImages int
}

func NewSession(username string) *Session {
    return &Session{
        username,
        make([]*image.Image, MAX_IMAGES),
        0,
    }
}

func getImage(url string) (image.Image, string, error) {
    fmt.Println("GET " + url)
    response, err := http.Get(url)
    if err != nil {
        panic(err)
    }
    defer response.Body.Close()
    return image.Decode(response.Body)
}

func ExtractImage(msg *tgbotapi.Message) (*BotImage, error) {
    fileId := ""

    // check if image exists, as a document or a compressed image
    if (*msg).Photo != nil {
        allPhotos := *(*msg).Photo
        photoSize := allPhotos[len(allPhotos) - 1]
        fileId = photoSize.FileID

    } else if msg.Document != nil {
        doc := (*msg.Document)
        if _, exists := ACCEPTED_MIMETYPE[doc.MimeType]; exists {
            fileId = doc.FileID
        }
    }

    if fileId != "" {
        // image exists, try to download it
        directLink, err := bot.GetFileDirectURL(fileId)
        if err != nil {
            return nil, err
        }
        img, f, err := getImage(directLink)
        if err != nil {
            return nil, err
        }
        return &BotImage{img, f, directLink}, nil
    }

    return nil, nil
}

var bot *tgbotapi.BotAPI
var sessions map[string]*Session = map[string]*Session{}

func Generate(chatId int64, session *Session) {
    images := session.Images[:session.NbImages]
    fmt.Println(len(images))
    data, err := combiner.DefaultCompose(images)
    if err != nil {
        fmt.Println(err)
        bot.Send(tgbotapi.NewMessage(chatId, fmt.Sprintf("Error generating image: %s", err)))
    }else {
        fmt.Println("SENDING IMAGE")
        b := tgbotapi.FileBytes{Name: "image.png", Bytes: data}
        _, err = bot.Send(tgbotapi.NewPhotoUpload(chatId, b))
        fmt.Println("... ERROR ", err)
    }
}

func HandleMessage(message *tgbotapi.Message) {
    if message == nil {
        return
    }

    log.Printf("UPDATE HOOK: [%s] %s\n", message.From.UserName, message)
    var response tgbotapi.MessageConfig

    if session, exists := sessions[message.From.UserName]; exists {
        log.Printf("GOT EXISTING SESSION : %s %d\n", session.UserName, session.NbImages)
        if message.Text == CMD_ABORT {
            delete(sessions, message.From.UserName)
            response = tgbotapi.NewMessage(message.Chat.ID, "aborted. See you soon!")

        } else if message.Text == CMD_GENERATE {
            response = tgbotapi.NewMessage(message.Chat.ID, "generating...\nPlease, wait.")
            delete(sessions, message.From.UserName)
            go Generate(message.Chat.ID, session)

        } else if message.Text == CMD_START {
            response = tgbotapi.NewMessage(message.Chat.ID,
                fmt.Sprintf("You already have submitted %d images. To start again, use %s first.", session.NbImages, CMD_ABORT))
        } else {
            img, err := ExtractImage(message)
            if err != nil {
                response = tgbotapi.NewMessage(message.Chat.ID,
                    fmt.Sprintf("Error extracting image (%s)", err))
            } else if img == nil {
                response = tgbotapi.NewMessage(message.Chat.ID,
                    fmt.Sprintf("Please, send an image or abort with %s", CMD_ABORT))
            } else {
                if session.NbImages > MAX_IMAGES {
                    response = tgbotapi.NewMessage(message.Chat.ID,
                        fmt.Sprintf("You reached the maximum number of images (%d)...", MAX_IMAGES))
                } else {
                    session.Images[session.NbImages] = &img.Img
                    session.NbImages++
                    fmt.Printf("Image ADDED : %s %s (idx: %d)\n", img.DirectLink, img.Img.Bounds(), session.NbImages)
                    response = tgbotapi.NewMessage(message.Chat.ID,
                        fmt.Sprintf("Image submitted sucessfully.\nYou have now %d images.", session.NbImages))
                }
            }
        }
    } else {
        fmt.Println("NO SESSION")
        fmt.Println(sessions)
        if message.Text == CMD_START {
            session :=  NewSession(message.From.UserName)
            sessions[session.UserName] = session
            log.Println(sessions)
            response = tgbotapi.NewMessage(message.Chat.ID, "session created.\nready to receive images! ")
        } else {
            response = tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Use %s to begin a new session", CMD_START))
        }
    }
    bot.Send(response)

}

func main() {
    var err error
    bot, err = tgbotapi.NewBotAPI("227564652:AAFc7HdSDi_OhIdISU_8JqVpeXsHzmVRTck")

    if err != nil {
        log.Panic(err)
    }

    bot.Debug = false

    log.Printf("Authorized on account %s", bot.Self.UserName)

    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60

    updates, err := bot.GetUpdatesChan(u)

    for update := range updates {
        HandleMessage(update.Message)
    }

}

// https://api.telegram.org/file/bot227564652:AAFc7HdSDi_OhIdISU_8JqVpeXsHzmVRTck/photo/file_24.jpg
// https://api.telegram.org/bot227564652:AAFc7HdSDi_OhIdISU_8JqVpeXsHzmVRTck/getFile?file_id=
