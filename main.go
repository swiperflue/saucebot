// This is free and unencumbered software released into the public domain.

// Anyone is free to copy, modify, publish, use, compile, sell, or
// distribute this software, either in source code form or as a compiled
// binary, for any purpose, commercial or non-commercial, and by any
// means.

// In jurisdictions that recognize copyright laws, the author or authors
// of this software dedicate any and all copyright interest in the
// software to the public domain. We make this dedication for the benefit
// of the public at large and to the detriment of our heirs and
// successors. We intend this dedication to be an overt act of
// relinquishment in perpetuity of all present and future rights to this
// software under copyright law.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

// For more information, please refer to <http://unlicense.org/>

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	tb "gopkg.in/tucnak/telebot.v2"
	"html"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

const MaxGoogleResults = 4

type Config struct {
	TelegramToken          string `json:"TelegramToken"`
	SaucenaoToken          string `json:"SaucenaoToken"`
	UserAgent              string `json:"UserAgent"`
	GoogleSearchURL        string `json:"GoogleSearchURL"`
	SaucenaoSearchURL      string `json:"SaucenaoSearchURL"`
	SearchTermPrefix       string `json:"SearchTermPrefix"`
	SearchTermSuffix       string `json:"SearchTermSuffix"`
	GoogleResultLinkPrefix string `json:"GoogleResultLinkPrefix"`
	GoogleResultLinkSuffix string `json:"GoogleResultLinkSuffix"`
}

// type used parsed google results
// NOTE: SearchTerm is basically what appears next to "Possible
// related search" when you reverse search an image on google
type GoogleResult struct {
	SearchTerm string
	Results    [MaxGoogleResults]string
}

// type used to store saucenao.com API results
type SaucenaoResult struct {
	Results []struct {
		Header struct {
			Similarity float64 `json:",string"`
		}
		Data struct {
			Ext_Urls []string
			Title    string
		}
	}
}

// global configuration
var Conf Config

// This function loads the bot's configuration from a json file specified by
// it's path It takes and returns the Config or an error in case it failed
// to open the file or decode the json data.
func loadConfiguration(file string) (Config, error) {
	var c Config
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return c, errors.New("Failed to read configuration file: " + err.Error())
	}
	jp := json.NewDecoder(f)
	err = jp.Decode(&c)
	if err != nil {
		return c, errors.New("Failed to decode configuration file: " + err.Error())
	}
	return c, nil
}

// This function is just a simple abstraction over regex to avoid replicating
// code. It finds a string between two other strings (left and right)
// inside another string (src) and returns an array of size i of matches.
// Returns nil in case the string wasn't found.
func findBetweenPatterns(src, left, right string, i int) []string {
	reg := fmt.Sprintf("(%s)(.*?)(%s)", left, right)
	re, err := regexp.Compile(reg)
	if err != nil {
		return nil
	}
	matches := re.FindAllString(src, i)
	if matches == nil {
		return nil
	}
	for j := range matches {
		matches[j] = strings.TrimPrefix(strings.TrimSuffix(matches[j], right), left)
	}
	return matches
}

// This function makes HTTP requests, it can be used to make normal requests
// or upload files when given a POST method and non empty file name (file)
// and field name (name) parameters.
func request(method, url, file, name string, redirect bool) (*http.Response, error) {
	client := &http.Client{}
	var rb bytes.Buffer
	var mpw *multipart.Writer

	// if redirect is false, tell the client to not follow redirections
	if redirect == false {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// if a file is specified prepare the multipart message
	if file != "" {
		f, err := os.Open(file)
		if err != nil {
			return nil, errors.New("[Technical Difficulties]: Failed to open file tmp")
		}
		defer f.Close()

		rb = bytes.Buffer{}
		mpw = multipart.NewWriter(&rb)
		fw, err := mpw.CreateFormFile(name, file)
		if err != nil {
			return nil, errors.New("[Technical Difficulties]: Failed to create file form")
		}
		_, err = io.Copy(fw, f)
		if err != nil {
			return nil, errors.New("[Technical Difficulties]: Failed to copy file to form")
		}
		_ = mpw.WriteField("image_content", "")
		mpw.Close()
	}
	req, err := http.NewRequest(method, url, &rb)
	if err != nil {
		return nil, errors.New("[Technical Difficulties]: Failed to create HTTP request")
	}
	req.Header.Add("User-Agent", Conf.UserAgent)
	// in case we are uploading a file, specify the data type
	// (this is needed for google images to work)
	if file != "" {
		req.Header.Add("Content-Type", mpw.FormDataContentType())
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("[Technical Difficulties]: Failed to retrieve HTTP response")
	}
	return resp, nil
}

// Default sauce getter, it gets the sauce from Google images, it does that by:
// - making a post request to google uploading the image
// - retrieving a response that contains a search URL
// - making a GET request to this URL
// - retrieving a response containing the actual search results
// - parsing the search results for the search term and relevant links
// - returning a GoogleResult struct (object! whatever)
func get_sauce(file string) (*GoogleResult, error) {
	// upload image to google
	resp, err := request("POST", Conf.GoogleSearchURL, file, "encoded_image", false)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	// grab the search link
	link := resp.Header.Get("Location")
	if len(link) == 0 {
		return nil, errors.New("[Technical Difficulties]: Failed to retrieve Google search url")
	}
	link += "&hl=en"

	// request the actual search results
	resp, err = request("GET", link, "", "", true)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("[Technical Difficulties]: Failed to read Google response data")
	}
	resp.Body.Close()
	htmlRes := string(body)

	// parse the search results
	term := findBetweenPatterns(htmlRes, Conf.SearchTermPrefix, Conf.SearchTermSuffix, MaxGoogleResults)
	if term == nil {
		return nil, errors.New("Failed to find what this image is about")
	}
	r := &GoogleResult{}
	r.SearchTerm = term[0]
	idx := strings.LastIndex(htmlRes, "Pages that include matching images")
	if idx == -1 {
		return nil, errors.New("Couldn't find sauce :(")
	}
	htmlRes = string(htmlRes[idx:len(htmlRes)])
	matches := findBetweenPatterns(htmlRes, Conf.GoogleResultLinkPrefix, Conf.GoogleResultLinkSuffix, MaxGoogleResults)
	for i := 0; i < len(r.Results) && i < len(matches); i++ {
		r.Results[i] = html.UnescapeString(matches[i])
	}

	return r, nil
}

// sauce getter for anime, it gets the sauce from saucenao.com, it does that by:
// - making a GET request to saucenao API uploading the image
// - retrieving a json response containing the results
// - parsing the json results into a SaucenaoResult Struct
// - sorting the results by similarity
// - returning a SaucenaoResult struct (object! whatever)
func get_anime_sauce(file string) (*SaucenaoResult, error) {
	// make api request
	resp, err := request("POST", Conf.SaucenaoSearchURL+Conf.SaucenaoToken, file, "file", true)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("[Technical Difficulties] Failed to read Saucenao response body data")
	}
	resp.Body.Close()

	// parse the json response
	sauce := &SaucenaoResult{}
	err = json.Unmarshal(body, sauce)
	if err != nil {
		return nil, errors.New("[Technical Difficulties] Failed to process Saucenoe json response")
	}
	if len(sauce.Results) == 0 {
		return nil, errors.New("Couldn't find sauce :(")
	}

	// sort the results by similarity
	sort.Slice(sauce.Results[:], func(i, j int) bool {
		return sauce.Results[i].Header.Similarity > sauce.Results[j].Header.Similarity
	})

	return sauce, nil
}

// This function is used to save the photo the user requested its sauce,
// it also handles wrong bot usage.
func savePhotoAndGetMessage(bot *tb.Bot, m *tb.Message) (*tb.Message, error) {
	if m.FromGroup() == false {
		return nil, errors.New("This bot is only available for usage in groups right now")
	}
	if m.IsReply() == false {
		return nil, errors.New("Please reply to the message you want to get its sauce")
	}
	rt := m.ReplyTo
	if rt == nil {
		return nil, errors.New("Seems like privacy mode is enabled, I can't access the replied-to message")

	}
	p := rt.Photo
	if p != nil {
		pf := &p.File
		bot.Download(pf, "tmp")
		return rt, nil
	}
	s := rt.Sticker
	if s != nil {
		pf := &s.File
		bot.Download(pf, "tmp")
		return rt, nil
	}
	u := rt.Sender
	if u != nil {
		pfps, err := bot.ProfilePhotosOf(u)
		if len(pfps) == 0 || err != nil {
			return nil, errors.New("User have no profile pictures")
		}
		pf := &pfps[0].File
		bot.Download(pf, "tmp")
		return rt, nil
	}
	return nil, errors.New("Unexpected wrong usage!")
}

func main() {
	// load configuration
	var err error
	Conf, err = loadConfiguration("config.json")
	if err != nil {
		log.Fatalln(err)
		return
	}

	// create a log file

	logFile, err := os.OpenFile("saucebot.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open/create log file %s\n", err)
		return
	}
	defer logFile.Close()

	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	// create the bot
	bot, err := tb.NewBot(tb.Settings{
		Token:     Conf.TelegramToken,
		Poller:    &tb.LongPoller{Timeout: 10 * time.Second},
		ParseMode: tb.ModeHTML,
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	bot.Handle("/sauce", func(m *tb.Message) {
		// save the image
		rt, err := savePhotoAndGetMessage(bot, m)
		if err != nil {
			bot.Reply(m, err.Error())
			log.Printf("[%s]: %s\n", m.Sender.Username, err)
			return
		}

		// get the sauce
		log.Printf("user [%s - %d] requested sauce.\n", m.Sender.Username, m.Sender.ID)
		sauce, err := get_sauce("tmp")
		if err != nil {
			bot.Reply(rt, err.Error())
			log.Printf("[%s]: %s\n", m.Sender.Username, err)
			err = os.Remove("tmp")
			if err != nil {
				log.Println(err)
			}
			return
		}

		// format the reply (should probably be moved to a function)
		r := fmt.Sprintf("Google Images says this is: <b>%s</b>\n\n", sauce.SearchTerm)
		r += strings.Join(sauce.Results[:], "\n")

		// send the reply
		bot.Reply(rt, r)
		err = os.Remove("tmp")
		if err != nil {
			log.Println(err)
		}
	})

	bot.Handle("/animesauce", func(m *tb.Message) {
		// save the image
		rt, err := savePhotoAndGetMessage(bot, m)
		if err != nil {
			bot.Reply(m, err.Error())
			log.Printf("[%s]: %s\n", m.Sender.Username, err)
			return
		}

		// get the sauce
		log.Printf("user [%s - %d] requested sauce.\n", m.Sender.Username, m.Sender.ID)
		sauce, err := get_anime_sauce("tmp")
		if err != nil {
			bot.Reply(rt, err.Error())
			log.Printf("[%s]: %s\n", m.Sender.Username, err)
			err = os.Remove("tmp")
			if err != nil {
				log.Println(err)
			}
			return
		}
		var warning string
		// format the reply (should probably be moved to a function)
		if sauce.Results[0].Header.Similarity < 60 {
			warning += "<b>Similiraty is below 60%, results might be bad</b>"
		}
		titles := "<b>Titles:</b>\n"
		sources := "<b>Sources:</b>\n"
		for i := range sauce.Results {
			if i != 0 && sauce.Results[i-1].Data.Title != "" {
				titles += ", "
			}
			titles += sauce.Results[i].Data.Title
			sources += strings.Join(sauce.Results[i].Data.Ext_Urls[:], "\n")
			sources += "\n"
		}
		r := fmt.Sprintf("%s\n%s\n%s\n", warning, titles, sources)

		// send the reply
		bot.Reply(rt, r)
		err = os.Remove("tmp")
		if err != nil {
			log.Println(err)
		}
	})

	bot.Handle("based bot", func(m *tb.Message) {
		bot.Reply(m, "Thank you hooman sensei")
	})

	// run the bot
	bot.Start()
}
