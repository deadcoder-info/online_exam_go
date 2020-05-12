package examwebportal

import (
	"net/http"
	"online_exam_go/examwebportal/examTemplate"

	"github.com/jinzhu/gorm"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var coockieStore = sessions.NewCookieStore([]byte("something-very-secret"))

//RunWebPortal starts running the exam web portal
func RunWebPortal(addr string, db *gorm.DB) error {
	router := mux.NewRouter()

	// Handle Static Files
	fs := http.FileServer(http.Dir("./examwebportal/examTemplate/"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// Min Home
	router.Path("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		examTemplate.Home(false, w)
	})

	// Visitor Web Portal
	router.Path("/signup").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SignUp(w, r, db)
	})

	router.Path("/phoneverify").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		PhoneVerify(w, r, db)
	})

	router.Path("/verifycode").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		VerifyCode(w, r, db)
	})

	router.Path("/login").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Login(w, r, db)
	})

	// Dashboard Web Portal
	router.Path("/dashboard").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Dashboard(w, r, db)
	})

	router.Path("/createexam").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		CreateExam(w, r, db)
	})

	router.Path("/addquestion/{exam_id}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AddQuestion(w, r, db)
	})

	router.Path("/history").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		History(w, r, db)
	})

	router.Path("/examdetail/{exam_id}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ExamDetail(w, r, db)
	})

	router.Path("/takeexam").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		TakeExam(w, r, db)
	})

	router.Path("/takeexam/code/{exam_code}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		TakeExamCode(w, r, db)
	})

	http.Handle("/", router)
	return http.ListenAndServe(addr, nil)
}
