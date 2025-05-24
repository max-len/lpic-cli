# Warning
This code was generated with the help of an AI agent and is intended primarily for testing purposes. Most of the generated code is unmodified and may not be production-ready.

## Features
- **LPIC Exam Preparation:** Practice with real-like questions for LPIC certifications (LPIC-1, LPIC-2, CKA, etc.).
- **Question Scraper:** Scrape and collect certification questions from web sources using a configurable YAML file.
- **Encryption/Decryption:** Securely encrypt and decrypt question sets using AES encryption and gzip compression.
- **Offline Mode:** Work with downloaded and decrypted question sets in JSON format, no internet required after setup.
- **Interactive CLI Client:** Answer questions, view explanations, and track your progress in a terminal UI.
- **Test Modes:**
  - List available certifications and test sets
  - Start a test with a specific test set
  - Filter for only incorrect or unknown questions
  - Randomize question order
- **Progress Tracking:** Save and load your answered questions and progress using NutsDB.
- **Easy Setup:** Simple commands to fetch, decrypt, and start practicing.
- **Mouse usage:** Mouse usage possible

### Notes
Feel free to modify or add everything you need

### lpic-learner
This project is designed to assist you for the learning of the LPIC (Linux Professional Institute Certification) content.  

### Fetch and decrypt the data
You need a url and a token for that. Ask someone.

```
export AES_KEY=XXXXXXXXXXXX....XXX..XX.X.X.
make build-tools && ./bin/crypt -operation=decrypt -url=https://XXX.XYZ/XXXX/output.json.enc -output=/tmp/test.json
```

## Run the client
```
make build-client
```

### List available Certifications
```
./bin/client --dbfile=test.json --listCerts
```
### List available test sets
```
./bin/client --dbfile=test.json --certId=lpic1-101-500 --listTestSets
```
### Start a test with a testset
```
 ./bin/client --dbfile=test.json --certId=lpic1-101-500 --testsetId=full_test_6
```
### Start test only with yet incorrect or unknown questions
```
 ./bin/client --dbfile=test.json --certId=lpic1-101-500 --testsetId=full_test_6 --filterCorrect
```
 ### Start a test set with all questions in a random order
 ```
 ./bin/client --dbfile=test.json --certId=lpic1-101-500 -randomQuestions  --filterCorrect

```

## Run the scraper
To scrape and encrypt data:
```
go run cmd/scraper/main.go -cookies='language=en-gb; currency=USD; .....XXYYZZZ......................' -config=cmd/scraper/config/lpic11.yaml

# Create encrypted json file
make  build-tools &&  ./bin/crypt -output=/tmp/output.json.enc

# Descrypt the created json file
make  build-tools &&  ./bin/crypt -operation=decrypt -input=/tmp/output.json.enc  -output=/tmp/tests.json
```
