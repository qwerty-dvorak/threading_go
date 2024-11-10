package main

import (
    "bytes"
    "encoding/base64"
    "encoding/csv"
    "encoding/json"
    "fmt"
    "html/template"
    "io"
    "log"
    "mime/multipart"
    "net/smtp"
    "net/textproto"
    "os"
    //"strconv"
    "strings"
)

type Image struct {
    Name     string
    Bytes    []byte
    MimeType string
    CID      string
}

const (
    FROM     = "adheeshgarg0611@gmail.com"
    FROMNAME = "Adheesh Garg"
    SUBJECT  = "Test Email with MIME"
    SMTPHOST = "smtp.gmail.com"
    SMTPPORT = "587"
)

var PASSWORD string

func real() string {
    if PASSWORD != "" {
        return PASSWORD
    }
    data, err := os.ReadFile("pass.txt")
    if err != nil {
        log.Fatal(err)
    }
    PASSWORD = string(data)
    return PASSWORD
}

func main() {
    real()

    // Open and parse the JSON file
    jsonFile, err := os.Open("temp.json")
    if err != nil {
        log.Fatalf("Unable to open JSON file: %v", err)
    }
    defer jsonFile.Close()

    // Read the JSON content
    byteValue, _ := io.ReadAll(jsonFile)

    // Define a struct to hold the attachments
    type Attachments struct {
        Attachments map[string]struct {
            Data string `json:"data"`
        } `json:"attachments"`
    }

    var attachments Attachments

    // Unmarshal the JSON content
    err = json.Unmarshal(byteValue, &attachments)
    if err != nil {
        log.Fatalf("Error unmarshalling JSON: %v", err)
    }

    // Create an array of Image structs
    var images []Image
    for name, attachment := range attachments.Attachments {
        dataParts := strings.SplitN(attachment.Data, ",", 2)
        if len(dataParts) != 2 {
            continue
        }
        mimeType := strings.TrimSuffix(strings.TrimPrefix(dataParts[0], "data:"), ";base64")
        data, err := base64.StdEncoding.DecodeString(dataParts[1])
        if err != nil {
            log.Printf("Error decoding base64 data for %s: %v", name, err)
            continue
        }
        images = append(images, Image{
            Name:     name,
            Bytes:    data,
            MimeType: mimeType,
            CID:      name,
        })
    }

    // Open the CSV file
    file, err := os.Open("tester.csv")
    if err != nil {
        log.Fatalf("Unable to open CSV file: %v", err)
    }
    defer file.Close()

    // Read CSV data
    reader := csv.NewReader(file)
    // Parse the email template
    tmpl, err := template.ParseFiles("template.txt")
    if err != nil {
        log.Fatalf("Error parsing template: %v", err)
    }

    headers, err := reader.Read()
    if err != nil {
        panic(err)
    }

    var records []map[string]string

    for {
        recordData, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            panic(err)
        }

        data := make(map[string]string)
        for i, value := range recordData {
            data[headers[i]] = value
        }
        records = append(records, data)
    }


    for _, record := range records {
        var htmlBody bytes.Buffer
        writer := multipart.NewWriter(&htmlBody)

        htmlBody.WriteString("From: " + FROM + "\r\n")
        htmlBody.WriteString("To: " + record["Email"] + "\r\n")
        htmlBody.WriteString("Subject: " + SUBJECT + "\r\n")
        htmlBody.WriteString("MIME-Version: 1.0\r\n")
        htmlBody.WriteString("Content-Type: multipart/related; boundary=" + writer.Boundary() + "\r\n")
        htmlBody.WriteString("\r\n") // End of headers

        // Create the HTML part
        htmlHeaders := make(textproto.MIMEHeader)
        htmlHeaders.Set("Content-Type", "text/html; charset=\"UTF-8\"")
        htmlPart, err := writer.CreatePart(htmlHeaders)
        if err != nil {
            log.Fatalf("Error creating HTML part: %v", err)
        }

        // Execute the template with record data
        err = tmpl.Execute(htmlPart, record)
        if err != nil {
            log.Fatalf("Error executing template: %v", err)
        }

        // Attach images
        for _, img := range images {
            imagePartHeaders := textproto.MIMEHeader{}
            imagePartHeaders.Set("Content-Type", img.MimeType)
            imagePartHeaders.Set("Content-Transfer-Encoding", "base64")
            imagePartHeaders.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", img.Name))
            imagePartHeaders.Set("Content-ID", fmt.Sprintf("<%s>", img.CID))

            imagePart, err := writer.CreatePart(imagePartHeaders)
            if err != nil {
                log.Fatalf("Error creating image part: %v", err)
            }
            encodedImage := base64.StdEncoding.EncodeToString(img.Bytes)

            // Wrap the base64 content to match the 76 character line length limit
            maxLineLength := 76
            for i := 0; i < len(encodedImage); i += maxLineLength {
                end := i + maxLineLength
                if end > len(encodedImage) {
                    end = len(encodedImage)
                }
                imagePart.Write([]byte(encodedImage[i:end] + "\r\n"))
            }
        }

        writer.Close()
        // Send the email using htmlBody.Bytes()
        fmt.Printf("Prepared email for: %s\n", record["Email"])
        auth := smtp.PlainAuth("", FROM, PASSWORD, SMTPHOST)

        // Send the email
        err = smtp.SendMail(SMTPHOST+":"+SMTPPORT, auth, FROM, []string{record["Email"]}, htmlBody.Bytes())
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("Email sent to: %s\n", record["Email"])
    }
}

// func atoi(str string) int {
//     num, _ := strconv.Atoi(str)
//     return num
// }