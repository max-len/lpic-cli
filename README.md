# Warning
This code was generated with the help of an AI agent and is intended primarily for testing purposes. Most of the generated code is unmodified and may not be production-ready.

## App usage

### CLI Keystrokes & Navigation
The lpic-learner CLI app provides an interactive terminal interface for practicing LPIC exam questions. After starting the client, you can use the following keystrokes to navigate and interact with the questions:

- **q**: Quit the application
- **n**: Go to the next question
- **p**: Go to the previous question
- **Space** / **Enter**: Select or mark the current answer option
- **s**: Show the solution (mark all answers and show explanation)
- **e**: Show the explanation for the current question
- **Up/Down arrows**: Navigate between answer options
- **t**: Show statistics
- **h**: Show help

todo: 
- switching from tview to lipgloss
- add a key to mark a question as important, to bring it back later
- keystroke to jump to the first unanswered question


You can also use the mouse to select answers and interact with the UI.

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
