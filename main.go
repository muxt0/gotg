package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "os"
)

// Helper function to send HTTP request
func sendRequest(url, contentType string, body io.Reader) (*http.Response, error) {
    req, err := http.NewRequest("POST", url, body)
    if err != nil {
        return nil, fmt.Errorf("❌ Failed to create request: %w", err)
    }
    req.Header.Set("Content-Type", contentType)
    return http.DefaultClient.Do(req)
}

// Helper function to add a file to the form
func addFileToForm(writer *multipart.Writer, fieldName, filePath string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return fmt.Errorf("❌ Failed to open file: %w", err)
    }
    defer file.Close()

    part, err := writer.CreateFormFile(fieldName, file.Name())
    if err != nil {
        return fmt.Errorf("❌ Failed to create form field: %w", err)
    }
    _, err = io.Copy(part, file)
    return err
}

// Single function to send text, files, and photos
func sendToTelegram(token, chatID, sendType, content, thumbPath string) {
    var url, contentType string
    var body io.Reader
    buf := &bytes.Buffer{}
    writer := multipart.NewWriter(buf)

    // Determine URL and sending method based on message type
    switch sendType {
    case "text":
        url = fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
        data := map[string]string{"chat_id": chatID, "text": content}
        jsonData, _ := json.Marshal(data)
        body = bytes.NewBuffer(jsonData)
        contentType = "application/json"
    case "photo":
        url = fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", token)
        writer.WriteField("chat_id", chatID)

        // Add the photo
        if err := addFileToForm(writer, "photo", content); err != nil {
            fmt.Println(err)
            return
        }

        // Add the thumbnail (optional)
        if thumbPath != "" {
            if err := addFileToForm(writer, "thumb", thumbPath); err != nil {
                fmt.Println("⚠️ Warning: Failed to add thumbnail:", err)
            }
        }
        writer.Close()
        body = buf
        contentType = writer.FormDataContentType()
    case "file":
        url = fmt.Sprintf("https://api.telegram.org/bot%s/sendDocument", token)
        writer.WriteField("chat_id", chatID)

        // Add the file
        if err := addFileToForm(writer, "document", content); err != nil {
            fmt.Println(err)
            return
        }

        // Add the thumbnail (optional)
        if thumbPath != "" {
            if err := addFileToForm(writer, "thumb", thumbPath); err != nil {
                fmt.Println("⚠️ Warning: Failed to add thumbnail:", err)
            }
        }
        writer.Close()
        body = buf
        contentType = writer.FormDataContentType()
    default:
        fmt.Println("❌ Invalid send type. It should be 'text', 'file', or 'photo'")
        return
    }

    // Send the request
    response, err := sendRequest(url, contentType, body)
    if err != nil {
        fmt.Println("❌ Failed to send request:", err)
        return
    }
    defer response.Body.Close()

    // Process the response
    if response.StatusCode == http.StatusOK {
        fmt.Println("✅ Sent successfully!")
    } else {
        responseData, _ := io.ReadAll(response.Body)
        fmt.Printf("❌ Failed to send! Error: %s\n📩 Server response: %s\n", response.Status, string(responseData))
    }
}

// Main function
func main() {
    if len(os.Args) > 1 && os.Args[1] == "--help" {
        fmt.Println("⚙️ Tool usage instructions:")
        fmt.Println("  <TOKEN> <CHAT_ID> text \"Text message\"  - To send a text message.")
        fmt.Println("  <TOKEN> <CHAT_ID> file \"File path\"   - To send a file.")
        fmt.Println("  <TOKEN> <CHAT_ID> file \"File path\" \"Thumbnail path\" - To send a file with a thumbnail.")
        fmt.Println("  <TOKEN> <CHAT_ID> photo \"Photo path\" - To send a photo.")
        return
    }

    if len(os.Args) < 5 {
        fmt.Println("❌ Incorrect command usage. Use --help to learn more.")
        return
    }

    token := os.Args[1]
    chatID := os.Args[2]
    sendType := os.Args[3]
    content := os.Args[4]
    thumbPath := ""
    if sendType == "file" && len(os.Args) >= 6 {
        thumbPath = os.Args[5]
    }

    sendToTelegram(token, chatID, sendType, content, thumbPath)
}