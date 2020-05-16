package examwebportal

import (
	"net/http"
	"online_exam_go/databaselayer"
	"online_exam_go/examwebportal/examTemplate"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

func Dashboard(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	switch r.Method {
	case "GET":
		examTemplate.Panel(w)
	}
}

func CreateExam(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	session, _ := coockieStore.Get(r, ACCESSTOKENSESSION)
	activeExamSession, _ := coockieStore.Get(r, CURRENTEXAMCREATESESSION)

	switch r.Method {
	case "GET":
		examTemplate.CreateExam(w)
	case "POST":
		user := DecodeAccessToken(session.Values[ACCESSTOKEN].(string))
		db.Where("phone_no = ?", user.PhoneNo).Find(&user)
		private, _ := strconv.ParseBool(r.FormValue("private"))

		timeInt, _ := strconv.Atoi(r.FormValue("exam_time"))
		exam := databaselayer.Exam{
			UserID:   user.ID,
			Name:     r.FormValue("exam_name"),
			Time:     timeInt,
			Password: r.FormValue("exam_password"),
			Public:   !private,
		}
		db.Create(&exam)
		activeExamSession.Values[CURRENTEXAMCREATEINDEX] = exam.ID
		activeExamSession.Save(r, w)
		examID := strconv.FormatUint(uint64(exam.ID), 16)
		http.Redirect(w, r, "/addquestion/"+examID, http.StatusSeeOther)
	}
}

func AddQuestion(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	vars := mux.Vars(r)

	examID, err := strconv.Atoi(vars["exam_id"])

	if err != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}

	exam := databaselayer.Exam{}
	db.First(&exam, examID)

	session, _ := coockieStore.Get(r, ACCESSTOKENSESSION)
	user := DecodeAccessToken(session.Values[ACCESSTOKEN].(string))
	db.Where("phone_no = ?", user.PhoneNo).Find(&user)

	if exam.UserID != user.ID {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}

	switch r.Method {
	case "GET":
		examTemplate.AddQuestionToExam(w)
	case "POST":
		if r.FormValue("add_question_btn") != "" {
			q := databaselayer.Question{
				ExamID:   exam.ID,
				Question: r.FormValue("question"),
				Answer1:  r.FormValue("answer_1"),
				Answer2:  r.FormValue("answer_2"),
				Answer3:  r.FormValue("answer_3"),
				Answer4:  r.FormValue("answer_4"),
			}
			db.Create(&q)
			examTemplate.AddQuestionToExam(w)
		}
	}
}

func History(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	session, _ := coockieStore.Get(r, ACCESSTOKENSESSION)
	user := DecodeAccessToken(session.Values[ACCESSTOKEN].(string))
	db.Where("phone_no = ?", user.PhoneNo).Find(&user)
	var exams []databaselayer.Exam

	db.Model(&user).Related(&exams)

	switch r.Method {
	case "GET":
		examTemplate.History(exams, w)
	}
}

func ExamDetail(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	vars := mux.Vars(r)
	exam := databaselayer.Exam{}
	session, _ := coockieStore.Get(r, ACCESSTOKENSESSION)
	user := DecodeAccessToken(session.Values[ACCESSTOKEN].(string))
	db.Where("phone_no = ?", user.PhoneNo).Find(&user)
	db.First(&exam, vars["exam_id"])
	if exam.CreatedAt.IsZero() || exam.UserID != user.ID {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	questions := []databaselayer.Question{}
	db.Model(&exam).Related(&questions)

	switch r.Method {
	case "GET":
		examTemplate.ExamDetail(exam, exam.CreatedAt.Format("2006-01-02 15:04:05"), questions, w)
	}
}

func TakeExam(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	session, _ := coockieStore.Get(r, ACCESSTOKENSESSION)
	user := DecodeAccessToken(session.Values[ACCESSTOKEN].(string))

	db.Where("phone_no = ?", user.PhoneNo).Find(&user)

	if user.ID == 0 {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	switch r.Method {
	case "GET":
		examTemplate.TakeExam(false, false, w)
	case "POST":
		examID := r.FormValue("exam_code")
		examPasword := r.FormValue("password")

		examParticipation := databaselayer.ExamParticipation{}

		db.Where("user_id = ? AND exam_id = ?", user.ID, examID).Find(&examParticipation)

		if examParticipation.ID > 0 {
			examTemplate.TakeExam(false, true, w)
			return
		}

		exam := databaselayer.Exam{}
		db.First(&exam, examID)

		var allowed bool = false

		if exam.ID > 0 {
			allowed = true
		}

		if exam.Password != "" {
			if exam.Password != examPasword {
				allowed = false
			}
		}

		if allowed {
			session, _ := coockieStore.Get(r, ALLOWEDEXAMSESSION)
			token, _ := CreateAllowedExamToken(exam.ID)

			endTime := time.Now().Local().Add(time.Hour*time.Duration(0) +
				time.Minute*time.Duration(exam.Time) +
				time.Second*time.Duration(0))

			examParticipation := databaselayer.ExamParticipation{
				UserID:  user.ID,
				ExamID:  uint(exam.ID),
				EndTime: endTime,
			}

			db.Save(&examParticipation)

			session.Values[ALLOWEDEXAMSESSIONID] = token
			session.Save(r, w)

			currentQuestionSession, _ := coockieStore.Get(r, CURRENTQUESTIONSESSION)
			currentQuestionSession.Values[CURRENTQUESTIONSESSIONID] = 0
			currentQuestionSession.Save(r, w)
			http.Redirect(w, r, "/takeexam/code/"+examID, http.StatusSeeOther)
			return
		}

		examTemplate.TakeExam(true, false, w)

	}
}

func TakeExamCode(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	session, _ := coockieStore.Get(r, ACCESSTOKENSESSION)
	user := DecodeAccessToken(session.Values[ACCESSTOKEN].(string))

	db.Where("phone_no = ?", user.PhoneNo).Find(&user)

	if user.ID == 0 {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	allowedExamSession, _ := coockieStore.Get(r, ALLOWEDEXAMSESSION)
	allowedExam := DecodeAllowedExam(allowedExamSession.Values[ALLOWEDEXAMSESSIONID].(string))

	vars := mux.Vars(r)
	examID := vars["exam_code"]

	examIDUint, _ := strconv.ParseUint(examID, 10, 32)

	if uint(examIDUint) != allowedExam {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}

	exam := databaselayer.Exam{}

	db.First(&exam, examID)

	examParticipation := databaselayer.ExamParticipation{}

	db.Where("user_id = ? AND exam_id = ?", user.ID, examID).Find(&examParticipation)

	x := examParticipation.EndTime.Format(time.RFC3339)

	currentQuestionSession, _ := coockieStore.Get(r, CURRENTQUESTIONSESSION)
	questionID := currentQuestionSession.Values[CURRENTQUESTIONSESSIONID]

	var questionIDUint uint = 0

	if questionID == nil || questionID == "" {
		questionIDUint = 0
	} else {
		questionIDUint64 := uint64(questionID.(int))
		questionIDUint = uint(questionIDUint64)
	}

	questions := []databaselayer.Question{}

	db.Model(&exam).Related(&questions)

	switch r.Method {
	case "GET":
		if time.Now().After(examParticipation.EndTime) {
			examTemplate.TakeExamCode(false, x, true, databaselayer.Question{}, false, w)
		}

		// if questionIDUint == 0 {
		// 	questionIDUint = questions[0].ID
		// }

		question := questions[questionIDUint]

		if question.ExamID != uint(examIDUint) {
			examTemplate.TakeExamCode(true, x, false, databaselayer.Question{}, false, w)
			return
		}

		examTemplate.TakeExamCode(false, x, false, question, false, w)
		currentQuestionSession.Values[CURRENTQUESTIONSESSIONID] = int(questionIDUint)
		currentQuestionSession.Save(r, w)
	case "POST":
		question := questions[questionIDUint]
		res := databaselayer.Result{
			UserID:      user.ID,
			QuestionID:  question.ID,
			Answer:      r.FormValue("result"),
			Description: r.FormValue("description"),
		}

		db.Save(&res)

		if int(questionIDUint) >= len(questions)-1 {
			examTemplate.TakeExamCode(false, x, false, databaselayer.Question{}, true, w)
			return
		}

		question = questions[questionIDUint+1]

		currentQuestionSession.Values[CURRENTQUESTIONSESSIONID] = int(questionIDUint + 1)
		currentQuestionSession.Save(r, w)

		examTemplate.TakeExamCode(false, x, false, question, false, w)

		// if questionIDUint == 0 {
		// 	questionIDUint = questions[0].ID
		// }

		// i := 1

		// for _, q := range questions {
		// 	if q.ID == questionIDUint {
		// 		if i+1 >= len(questions) {
		// 			token, _ := CreateAllowedExamToken(0)
		// 			allowedExamSession.Values[ALLOWEDEXAMSESSIONID] = token
		// 			allowedExamSession.Save(r, w)
		// 			examTemplate.TakeExamCode(false, x, false, databaselayer.Question{}, true, w)
		// 			return
		// 		}
		// 		currentQuestionSession.Values[CURRENTQUESTIONSESSIONID] = i + 1
		// 		currentQuestionSession.Save(r, w)
		// 		question := questions[i+1]
		// 		examTemplate.TakeExamCode(false, x, false, question, false, w)
		// 	}
		// 	i++
		// }
	}
}
