package examwebportal

import (
	"math/rand"
	"net/http"
	"online_exam_go/databaselayer"
	"online_exam_go/examwebportal/examTemplate"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
)

const (
	INPUTNAME                = "inputname"
	SMSTOKENSESSION          = "smstokensession"
	ACCESSTOKENSESSION       = "accesstokensession"
	SMSTOKEN                 = "smstoken"
	ACCESSTOKEN              = "access_token"
	CURRENTEXAMCREATESESSION = "currentexamcreate"
	CURRENTEXAMCREATEINDEX   = "cureentcreateexam"
	ALLOWEDEXAMSESSION       = "allowedexamsession"
	ALLOWEDEXAMSESSIONID     = "allowedsessionid"
	CURRENTQUESTIONSESSION   = "currentquestionsession"
	CURRENTQUESTIONSESSIONID = "currentquestionsessionid"
)

func SignUp(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	switch r.Method {
	case "GET":
		examTemplate.SignUp(false, false, w)
	case "POST":
		session, _ := coockieStore.Get(r, SMSTOKENSESSION)
		_, phoneNo, validated := DecodeSMSToken(session.Values[SMSTOKEN].(string))
		if !validated {
			delete(session.Values, SMSTOKEN)
			session.Save(r, w)
			panic("Invalid token")
		}

		if r.FormValue("password") != r.FormValue("password_confirm") {
			examTemplate.SignUp(false, true, w)
			return
		}

		var userCount int = 0
		db.Model(&databaselayer.User{}).Where(&databaselayer.User{UserName: r.FormValue("username")}).Count(&userCount)

		if userCount >= 1 {
			examTemplate.SignUp(true, false, w)
			return
		}

		u := databaselayer.User{
			PhoneNo:  phoneNo,
			UserName: r.FormValue("username"),
			Password: r.FormValue("password"),
			UserType: "normal",
		}
		db.Create(&u)
		session.Options.MaxAge = -1
		session.Save(r, w)
		http.Redirect(w, r, "/login/success", http.StatusSeeOther)
	}
}

func PhoneVerify(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	switch r.Method {
	case "GET":
		examTemplate.PhoneVerify(false, w)
	case "POST":
		phoneNo := r.FormValue("phone_no")
		a := databaselayer.User{
			PhoneNo: phoneNo,
		}
		db.Where(&a).First(&a)

		if a.CreatedAt.IsZero() {
			rand.Seed(time.Now().UnixNano())
			code := int64(rand.Intn(899999))
			code = code + 100000

			strCode := strconv.Itoa(int(code))

			SendKavenegarOneToken("shakiba-exam-verification", phoneNo, strCode)

			token, _ := CreateSMSCodeToken(phoneNo, code, false)

			session, _ := coockieStore.Get(r, SMSTOKENSESSION)

			session.Values[SMSTOKEN] = token
			session.Save(r, w)

			http.Redirect(w, r, "/verifycode", http.StatusSeeOther)
			return
		}

		examTemplate.PhoneVerify(true, w)
	}
}

func VerifyCode(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	switch r.Method {
	case "GET":
		examTemplate.VerifyCode(false, w)
	case "POST":
		session, _ := coockieStore.Get(r, SMSTOKENSESSION)
		sessionCode, sessionPhoneNo, _ := DecodeSMSToken(session.Values[SMSTOKEN].(string))
		userCode, _ := strconv.ParseInt(r.FormValue("code"), 10, 64)
		if sessionCode != userCode {
			examTemplate.VerifyCode(true, w)
			return
		}
		token, _ := CreateSMSCodeToken(sessionPhoneNo, sessionCode, true)

		session.Values[SMSTOKEN] = token
		session.Save(r, w)

		http.Redirect(w, r, "/signup", http.StatusSeeOther)
	}
}

func Login(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	session, _ := coockieStore.Get(r, ACCESSTOKENSESSION)
	if session.Values[ACCESSTOKEN] != nil && session.Values[ACCESSTOKEN] != "" {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}

	switch r.Method {
	case "GET":
		examTemplate.Login(false, false, w)
	case "POST":
		userName := r.FormValue("username")
		password := r.FormValue("password")

		u := databaselayer.User{}

		db.Where(&databaselayer.User{UserName: userName}).First(&u)

		if u.CreatedAt.IsZero() || u.Password != password {
			examTemplate.Login(true, false, w)
			return
		}

		token, _ := CreateAccessToken(u)
		session.Values[ACCESSTOKEN] = token
		session.Save(r, w)

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}
