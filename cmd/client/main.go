package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/SqiSch/lpic-cli/internal/database"
	"github.com/SqiSch/lpic-cli/internal/repository"
	"github.com/SqiSch/lpic-cli/internal/types"
	"github.com/SqiSch/lpic-cli/internal/views"
)

func fetchQuestionByIndex(questions []*types.Question, index int) (*types.Question, error) {
	log.Println("fetchQuestionByIndex", index)
	if index < 0 || index >= len(questions) {
		return questions[len(questions)-1], nil
	}

	keys := make([]int, 0, len(questions))
	for key := range questions {
		keys = append(keys, key)
	}

	questionID := keys[index]
	return questions[questionID], nil
}

func main() {

	var rep repository.QuestionRepository = repository.NewNutsQuestionRepository()
	ctx := context.Background()

	// Add a flag for the database filename
	dbFile := flag.String("dbfile", "test.json", "Path to the JSON database file containing certification sets")
	certID := flag.String("certId", "lpic1-101-500", "Id of the certification set to load from the json file")
	testSetId := flag.String("testsetId", "admin_1", "Id of the test set to load from the json file")
	listCerts := flag.Bool("listCerts", false, "List all available certifications in the json file")
	listTestSets := flag.Bool("listTestSets", false, "List all available test sets in the json file")
	filterCorrect := flag.Bool("filterCorrect", false, "Filter correct answers")
	withLogfile := flag.Bool("withLogfile", false, "Enable logging to a file in /tmp/lpic-learner.log")
	help := flag.Bool("help", false, "Show help")
	h := flag.Bool("h", false, "Show help")
	randomQuestions := flag.Bool("randomQuestions", false, "Fetch random questions from the certification set instead of a specific test set")
	flag.Parse()

	if *help || *h {
		fmt.Println("Usage: lpic-learner [options]")
		fmt.Println("Options:")
		fmt.Println("  -dbfile string")
		fmt.Println("        Path to the JSON database file containing certification sets (default \"test.json\")")
		fmt.Println("  -certId string")
		fmt.Println("        Id of the certification set to load from the json file (default \"lpic1-101-500\")")
		fmt.Println("  -testsetId string")
		fmt.Println("        Id of the test set to load from the json file (default \"admin_1\")")
		fmt.Println("  -listCerts")
		fmt.Println("        List all available certifications in the json file")
		fmt.Println("  -listTestSets")
		fmt.Println("        List all available test sets in the json file")
		fmt.Println("  -randomQuestions")
		fmt.Println("        Fetch random questions from the certification set instead of a specific test set")
		fmt.Println("        If this option is set, the -testsetId option is ignored")
		fmt.Println("  -filterCorrect")
		fmt.Println("        Filter correct answers")
		fmt.Println("  -withLogfile")
		fmt.Println("        Enable logging to a file in /tmp/lpic-learner.log")
		fmt.Println("        If this option is set, the log file is created in /tmp/lpic-learner.log")
		fmt.Println("  -help")
		fmt.Println("        Show help")
		fmt.Println("Examples:")
		fmt.Println("  lpic-learner -dbfile test.json -certId lpic1-101-500 -testsetId admin_1")
		fmt.Println("  lpic-learner -listCerts")
		fmt.Println("  lpic-learner -listTestSets -certId lpic1-101-500")
		fmt.Println("  lpic-learner--dbfile=test.json --certId=lpic1-101-500 --testsetId=admin_1 --filterCorrect")
		return
	}

	if *withLogfile {
		LOG_FILE := "/tmp/lpic-learner.log"
		log.Println("Logging enabled to file:", LOG_FILE)
		logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Panic(err)
		}
		defer logFile.Close()
		log.SetOutput(logFile)
		log.SetFlags(log.Lshortfile | log.LstdFlags)
	} else {
		log.SetOutput(ioutil.Discard)
		log.SetFlags(0)
	}

	certSet, err := database.LoadDatabaseFromFile(*dbFile, *certID)
	if err != nil {
		log.Fatalf("failed to load certification set: %v", err)
	}

	log.Printf("Loaded certification set: %s (%s) %d questions \n", certSet.CertificationName, certSet.CertificationID, len(certSet.Questions))

	if *listCerts {
		fmt.Println("Available certifications:")
		certs, err := database.LoadFullData(*dbFile)
		if err != nil {
			log.Fatalf("failed to load certification sets: %v", err)
		}
		for _, cert := range certs {
			fmt.Printf("ID: %s, Name: %s\n", cert.CertificationID, cert.CertificationName)
		}
		return
	}

	if *listTestSets {
		fmt.Println("Available test sets:")
		if certID == nil {
			log.Fatalf("certID is required to list test sets")
		}
		certSet, err := database.LoadDatabaseFromFile(*dbFile, *certID)
		if err != nil {
			log.Fatalf("failed to load certification set: %v", err)
		}
		for _, testSet := range certSet.Testsets {
			fmt.Printf("ID: %s, Name: %s, Numer of Questions: %d \n", testSet.TestsetID, testSet.TestsetName, len(testSet.QuestionsIds))
		}
		return
	}

	// Get the state of answered questions
	formerQuestionStates, err := rep.GetAnsweredQuestions()
	if err != nil {
		log.Printf("No answered questions found for certID %s: %v", *certID, err)
	}

	app := tview.NewApplication()

	var questions []*types.Question
	if *randomQuestions {
		questions, err = certSet.GetQuestionsForTestset("", *filterCorrect, formerQuestionStates)
		if err != nil {
			log.Fatalf("failed to fetch question: %v", err)
		}

		// create a fake testset for random questions
		testSet := types.Testset{
			TestsetID:    "random",
			TestsetName:  "Random Questions",
			QuestionsIds: make([]int, 0),
		}
		for _, question := range questions {
			testSet.QuestionsIds = append(testSet.QuestionsIds, question.ID)
		}
		certSet.Testsets["random"] = testSet
		testSetId = &testSet.TestsetID
	} else {
		questions, err = certSet.GetQuestionsForTestset(*testSetId, *filterCorrect, formerQuestionStates)
		if err != nil {
			log.Fatalf("failed to fetch question: %v", err)
		}
	}

	testset := certSet.Testsets[*testSetId]
	session := types.NewCertificationSession(&testset)

	question, err := fetchQuestionByIndex(questions, session.GetFirst())
	if err != nil {
		log.Fatalf("failed to fetch question: %v", err)
	}

	questionTextView := tview.NewTextView().SetText(question.Text).SetDynamicColors(true)

	explainationView := tview.NewTextView().SetText("").SetDynamicColors(true)

	flexViewflex := tview.NewFlex().SetDirection(tview.FlexRow)
	flexViewflex.AddItem(questionTextView, 1, 1, false)
	questionView := views.NewQuestionsView(question.Answers, questionTextView, explainationView)
	questionView.SetBorder(true).
		SetTitle("Answers").
		SetRect(0, 0, 30, 5)
	questionView.SetQuestion(question)
	flexViewflex.AddItem(questionView, 0, 1, true)
	flexViewflex.AddItem(explainationView, 0, 1, false)

	frame2 := tview.NewFlex().SetDirection(tview.FlexRow)
	frame2.AddItem(tview.NewButton("Explain"), 1, 0, false)
	frame2.AddItem(tview.NewButton("Solve"), 1, 0, false)
	frame2.AddItem(tview.NewButton("Next"), 1, 0, false)

	flex := tview.NewFlex()

	flexBoxTest := tview.NewFlex().SetDirection(tview.FlexRow)

	textcieTest := tview.NewTextView().SetText("").SetDynamicColors(true).SetTextAlign(tview.AlignCenter). // Center text within its area
														SetWrap(true).
														SetDynamicColors(true)
	views.QuestionStateOverview(questions, textcieTest, session.GetCurrentQuestionIndex())

	boxTest := tview.NewBox().SetBorder(true).SetTitle("")
	grid := tview.NewGrid().
		SetRows(0).       // Single row, expands vertically
		SetColumns(0).    // Single column, expands horizontally
		SetBorders(false) // The outer grid itself doesn't need borders
	grid.AddItem(textcieTest,
		0,    // row 0 (same as box)
		0,    // column 0 (same as box)
		1,    // row span 1 (same as box)
		1,    // column span 1 (same as box)
		0,    // min height
		0,    // min width
		true, //
	)

	flexBoxTest.AddItem(boxTest, 0, 1, false)
	flexBoxTest.AddItem(textcieTest, 0, 1, false)

	flex.AddItem(tview.NewBox().SetBorder(true).SetTitle("XXYYZZ"), 1, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(questionTextView, 0, 1, false).
			AddItem(questionView, 0, 3, true).
			AddItem(explainationView, 5, 1, false), 0, 2, false).
		AddItem(grid, 20, 1, false)

	modal := tview.NewModal()

	prevQuestion := func() {
		question, err = fetchQuestionByIndex(questions, session.GetAndDecIndex())
		if err != nil {
			log.Fatalf("failed to fetch question: %v", err)
		}
		questionView.SetQuestion(question)
		views.QuestionStateOverview(questions, textcieTest, session.GetCurrentQuestionIndex())
	}

	nextQuestion := func() {
		question, err = fetchQuestionByIndex(questions, session.GetAndIncIndex())
		if err != nil {
			log.Fatalf("failed to fetch question: %v", err)
		}
		questionView.SetQuestion(question)
		views.QuestionStateOverview(questions, textcieTest, session.GetCurrentQuestionIndex())
	}

	showStatistics := func() {
		testSetQuestionLenght := len(certSet.Questions)
		questionsLenght := len(questions)
		AlreadyAnswered := len(formerQuestionStates)

		CorrectAnswered := 0
		IncorrectAnswered := 0
		for _, question := range certSet.Questions {
			if question.AnsweredState == types.AnsweredTrue {
				CorrectAnswered++
			} else if question.AnsweredState == types.AnsweredFalse {
				IncorrectAnswered++
			}
		}

		CorrectAnsweredTotal := 0
		IncorrectAnsweredTotal := 0
		for _, question := range certSet.Questions {
			if question.AnsweredState == types.AnsweredTrue {
				CorrectAnsweredTotal++
			} else if question.AnsweredState == types.AnsweredFalse {
				IncorrectAnsweredTotal++
			}
		}

		precentageCorrect := float64(CorrectAnswered) / float64(questionsLenght) * 100
		precentageIncorrect := float64(IncorrectAnswered) / float64(questionsLenght) * 100
		precentageCorrectTotal := float64(CorrectAnsweredTotal) / float64(testSetQuestionLenght) * 100
		precentageIncorrectTotal := float64(IncorrectAnsweredTotal) / float64(testSetQuestionLenght) * 100

		questionCount := fmt.Sprintf("Testset: %s\nQuestions: %d\nAnswered: %d\nCorrect: %d\nIncorrect: %d\n", certSet.Testsets[*testSetId].TestsetName, testSetQuestionLenght, AlreadyAnswered, CorrectAnsweredTotal, IncorrectAnsweredTotal)
		questionCount += fmt.Sprintf("Correct Total: %.2f%%\nIncorrect: %.2f%%\n", precentageCorrectTotal, precentageIncorrectTotal)
		questionCount += fmt.Sprintf("Correct Testset: %.2f%%\nIncorrect: %.2f%%\n", precentageCorrect, precentageIncorrect)

		modal = tview.NewModal().
			SetText(questionCount).
			AddButtons([]string{"Ok"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if err := app.SetRoot(flex, false).EnableMouse(true).Run(); err != nil {
					panic(err)
				}
			})
		if err := app.SetRoot(modal, false).EnableMouse(true).Run(); err != nil {
			panic(err)
		}
	}

	markAnswer := func() {
		state := questionView.ToggleCurrentMarkedOption()
		question.SetAnsweredState(state)
		views.QuestionStateOverview(questions, textcieTest, session.GetCurrentQuestionIndex())
		rep.UpsertQuestion(ctx, question)
	}

	showHelp := func() {
		helpText := "Help:\n" +

			"q: Quit\n" +
			"Enter: Mark answer\n" +
			"Space: Mark answer\n" +
			"n/right: Next question\n" +
			"p/left: Previous question\n" +
			"t: Show statistics\n" +
			"e: Show explanation\n" +
			"h: Show help\n" +
			"s: Solve question\n"
		modal = tview.NewModal().
			SetText(helpText).
			AddButtons([]string{"Ok"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if err := app.SetRoot(flex, false).EnableMouse(true).Run(); err != nil {
					panic(err)
				}
			})
		if err := app.SetRoot(modal, false).EnableMouse(true).Run(); err != nil {
			panic(err)
		}
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				app.Stop()
			case ' ':
				markAnswer()
			case 'h':
				showHelp()
			case 'n':
				nextQuestion()
			case 'p':
				prevQuestion()
			case 's':
				// Solve the questions
				for _, v := range questionView.GetCurrentQuestion().Answers {
					v.SetIsMarked(true)
					questionView.GetCurrentQuestion().SetAnsweredState(types.AnsweredFalse)
				}
				questionView.ShowExplanation()
				views.QuestionStateOverview(questions, textcieTest, session.GetCurrentQuestionIndex())

			case 't':
				showStatistics()
			case 'e':
				explainationView.SetText(question.Explanation)
				log.Println("Show explanation")
			}
		case tcell.KeyEnter:
			markAnswer()
		case tcell.KeyUp:
			questionView.DecreaseMarkerPosition()
		case tcell.KeyLeft:
			prevQuestion()
		case tcell.KeyRight:
			nextQuestion()
		case tcell.KeyDown:
			questionView.IncreaseMarkerPosition()

		}
		return event
	})

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
