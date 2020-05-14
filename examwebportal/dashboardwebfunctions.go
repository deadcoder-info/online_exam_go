package examwebportal

import (
	"fmt"
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
		private, _ := strconv.ParseBool(r.FormValue("private"))
		fmt.Println(activeExamSession.Values[CURRENTEXAMCREATEINDEX])
		exam := databaselayer.Exam{
			UserID:   user.ID,
			Name:     r.FormValue("exam_name"),
			Time:     r.FormValue("exam_time"),
			Password: r.FormValue("exam_password"),
			Public:   !private,
		}
		db.Create(&exam)
		activeExamSession.Values[CURRENTEXAMCREATEINDEX] = exam.ID
		activeExamSession.Save(r, w)
		http.Redirect(w, r, "/addquestion", http.StatusSeeOther)
	}
}

func AddQuestion(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	activeExamSession, _ := coockieStore.Get(r, CURRENTEXAMCREATESESSION)

	vars := mux.Vars(r)

	examID, err := strconv.Atoi(vars["exam_id"])

	if err != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}

	exam := databaselayer.Exam{}
	db.First(&exam, examID)

	session, _ := coockieStore.Get(r, ACCESSTOKENSESSION)
	user := DecodeAccessToken(session.Values[ACCESSTOKEN].(string))

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
		if r.FormValue("submit_exam") != "" {
			activeExamSession.Values[CURRENTEXAMCREATEINDEX] = ""
			activeExamSession.Save(r, w)
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		}
	}
}

func History(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	session, _ := coockieStore.Get(r, ACCESSTOKENSESSION)
	user := DecodeAccessToken(session.Values[ACCESSTOKEN].(string))
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
		examTemplate.TakeExam(false, w)
	case "POST":
		examID := r.FormValue("exam_code")
		examPasword := r.FormValue("password")

		examParticipation := databaselayer.ExamParticipation{}

		db.Where("UserID = ? AND ExamID = ?", user.ID, examID).Find(&examParticipation)

		if examParticipation.ID >= 0 {
			examTemplate.TakeExam(false, true, w)
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
			session.Values[ALLOWEDEXAMSESSIONID] = token
			session.Save(r, w)
			http.Redirect(w, r, "/takeexam/code/"+examID, http.StatusSeeOther)
			return
		}

		examTemplate.TakeExam(true, w)

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

	switch r.Method {
	case "GET":
		timein := time.Now().Local().Add(time.Hour*time.Duration(0) +
			time.Minute*time.Duration(70) +
			time.Second*time.Duration(0))
		fmt.Println(timein.Format(time.RFC3339))
		x := timein.Format(time.RFC3339)
		examTemplate.TakeExamCode(false, x, w)
	case "POST":
		examID := r.FormValue("exam_code")
		examPasword := r.FormValue("password")

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
			session.Values[ALLOWEDEXAMSESSIONID] = token
			session.Save(r, w)
			http.Redirect(w, r, "/takeexam/code/"+examID, http.StatusSeeOther)
			return
		}

		examTemplate.TakeExam(true, w)

	}
}
