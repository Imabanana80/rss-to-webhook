package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mmcdole/gofeed"
	"gopkg.in/yaml.v2"
)

type Feed struct {
    Last string `yaml:"last"`
    Hook string `yaml:"hook"`
}

type Config struct {
    Feeds map[string]Feed `yaml:"feeds"`
}

func main() {
    for {
        go sendFeeds()
        time.Sleep(time.Minute)
    }    
}

func sendFeeds() {
    log.Println("Sending updated feeds...")
    data, err := os.ReadFile("config.yml")
    if err != nil {
        log.Fatal(err)
    }
    var config Config
    err = yaml.Unmarshal(data, &config)
    if (err != nil) {
        log.Fatal(err)
    }

    for name, feed := range config.Feeds {
        last := feed.Last
        feedURL := name    
        parser := gofeed.NewParser()

        feedResult, err := parser.ParseURL(feedURL)
        result := feedResult.Items[0] 
        if (err != nil) {
            log.Println(err)
        }
        if result.Title == last {
            continue
        }
        feed.Last = result.Title
        config.Feeds[name] = feed
        
        webhook := feed.Hook
        
        data := map[string]interface{}{
            "content": "## " + result.Title + "\n> " + result.Description +"\n\n-# " + "[Read...](" + result.Link + ")",
        }

        jsonData, err := json.Marshal(data)
        if err != nil {
            log.Fatal(err)
        }
        req, err := http.NewRequest("POST", webhook, bytes.NewBuffer(jsonData))
        if err != nil {
            log.Fatal(err)
        }
        req.Header.Set("Content-Type", "application/json")
        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
            log.Fatalf("error sending request: %v", err)
        }
        defer resp.Body.Close()
        log.Println("Posted: " + result.Title + " > " + result.Link)
    }

    updatedData, err :=  yaml.Marshal(&config)
    if err != nil {
        log.Fatal(err)
    }
    _ = os.WriteFile("config.yml", updatedData, 0644)

}
