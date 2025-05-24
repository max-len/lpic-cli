package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"io/ioutil"

	"github.com/SqiSch/lpic-cli/internal/types"
	"gopkg.in/yaml.v2"
)

type CertificationSetUrls struct {
	Title       string
	Description string
	Urls        map[string]string
}

type CertificationSetConfig struct {
	CertificationID          string `yaml:"certification_id"`
	CertificationName        string `yaml:"certification_name"`
	CertificationDescription string `yaml:"certification_description"`
	Urls                     []CertificationSetUrls
}

func readCertificationSetConfig(filePath string) (*CertificationSetConfig, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config CertificationSetConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func main() {
	// Command-line arguments
	cookies := flag.String("cookies", "", "Cookies to include in the request")
	configFile := flag.String("config", "lpic11.yaml", "Path to the YAML configuration file")
	flag.Parse()

	// Read the YAML configuration
	configFilePath := configFile
	certificationSetConfig, err := readCertificationSetConfig(*configFilePath)
	if err != nil {
		log.Fatalf("Failed to read certification set config: %v", err)
	}

	log.Printf("Loaded certification set config: %+v", certificationSetConfig)

	var certificationSet = types.CertificationSet{
		ID:                       certificationSetConfig.CertificationID,
		CertificationID:          certificationSetConfig.CertificationID,
		CertificationName:        certificationSetConfig.CertificationName,
		CertificationDescription: certificationSetConfig.CertificationDescription,
		Questions:                map[int]*types.Question{},  // all questions
		Testsets:                 map[string]types.Testset{}, // questions per testset
	}

	certificationSet.Questions = make(map[int]*types.Question)
	certificationSet.Testsets = make(map[string]types.Testset)

	// Get MongoDB credentials from environment variables
	mongoUser := os.Getenv("MONGO_USER")
	mongoPassword := os.Getenv("MONGO_PASSWORD")
	mongoURI := "mongodb://localhost:27017"
	if mongoUser != "" && mongoPassword != "" {
		mongoURI = "mongodb://" + mongoUser + ":" + mongoPassword + "@localhost:27017"
	}

	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.TODO())

	collection := client.Database("certificationDB").Collection("questions")

	// Regex to remove question prefix like "Q. 17  "
	questionPrefixRegex := regexp.MustCompile(`^Q\.\s*\d+\s+`)

	// Split URLs and scrape each
	for _, certsetConfig := range certificationSetConfig.Urls {
		for setId, urlstring := range certsetConfig.Urls {
			log.Printf("Scraping URL: %s", urlstring)

			// Create HTTP request with cookies
			req, err := http.NewRequest("GET", urlstring, nil)
			if err != nil {
				log.Printf("Failed to create request for URL %s: %v", urlstring, err)
				continue
			}
			if *cookies != "" {
				req.Header.Set("Cookie", *cookies)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Printf("Failed to fetch URL %s: %v", urlstring, err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Printf("Non-OK HTTP status for URL %s: %d", urlstring, resp.StatusCode)
				continue
			}

			// Parse the HTML
			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				log.Printf("Failed to parse HTML for URL %s: %v", urlstring, err)
				continue
			}

			// if url paramter contains type=fulllength then extract the testset param fromt the url
			urlObj, err := url.Parse(urlstring)
			if err != nil {
				log.Printf("Failed to parse URL %s: %v", urlstring, err)
				continue
			}

			queryParams := urlObj.Query()
			testset := queryParams.Get("testset")
			if testset == "" {
				log.Printf("No testset parameter found in URL %s", urlstring)
				continue
			}

			testType := queryParams.Get("type")
			if testType == "" {
				log.Printf("No type parameter found in URL %s", urlstring)
				continue
			}

			testsetObj := types.Testset{
				TestsetID:          setId,
				TestsetName:        certsetConfig.Title,
				TestsetDescription: certsetConfig.Description,
				QuestionsIds:       []int{},
			}

			// Extract questions and answers
			doc.Find(".card-group").Each(func(i int, s *goquery.Selection) {
				questionText := strings.TrimSpace(s.Find(".card-header h6").Text())
				questionText = questionPrefixRegex.ReplaceAllString(questionText, "")
				explanation := strings.TrimSpace(s.Find(".explanation").Text())
				explanation = strings.ReplaceAll(explanation, "Explanation:-  ", "")
				questionID, _ := s.Find("input[name^='question']").Attr("value")
				questionIDInt, _ := strconv.Atoi(questionID)

				var answers []*types.Answer
				s.Find(".card-content .radio").Each(func(j int, a *goquery.Selection) {
					answerText := strings.TrimSpace(a.Find("span").Text())
					isCorrect := a.Find("input").AttrOr("val", "0") == "1"
					answerID, _ := a.Find("input").Attr("value")

					answers = append(answers, &types.Answer{
						Text:      answerText,
						IsCorrect: isCorrect,
						AnswerID:  answerID,
					})
				})

				question := types.Question{
					ID:          questionIDInt,
					Text:        questionText,
					Answers:     answers,
					Explanation: explanation,
				}

				certificationSet.Questions[question.ID] = &question
				testsetObj.QuestionsIds = append(testsetObj.QuestionsIds, questionIDInt)
			})

			certificationSet.Testsets[setId] = testsetObj
		}
	}

	// Save to MongoDB
	filter := bson.M{"_id": certificationSetConfig.CertificationID}
	update := bson.M{"$set": certificationSet}
	_, err = collection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))
	if err != nil {
		log.Printf("Failed to update certification set: %v", err)
	} else {
		log.Printf("Updated certification set: %s", certificationSetConfig.CertificationID)
	}
}
