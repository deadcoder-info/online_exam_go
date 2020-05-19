package databaselayer

import (
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type User struct {
	gorm.Model
	UserType string
	PhoneNo  string `gorm:"type:varchar(15);unique_index"`
	UserName string
	Password string
	Time     int
	Exams    []Exam
	Users    []User
}

type Exam struct {
	gorm.Model
	UserID             uint
	Name               string
	Password           string
	Public             bool
	Time               int
	Questions          []Question
	ExamParticipations []ExamParticipation
}

type UserAllowedToExam struct {
	gorm.Model
	Exam    Exam
	PhoneNo string
}

type Question struct {
	gorm.Model
	ExamID     uint
	Question   string
	Answer1    string
	Answer2    string
	Answer3    string
	Answer4    string
	AnswerDesc string
	Results    []Result
}

type ExamParticipation struct {
	gorm.Model
	UserID  uint
	ExamID  uint
	EndTime time.Time
}

type Result struct {
	gorm.Model
	UserID      uint
	QuestionID  uint
	Answer      string
	Description string
}

func NewDataBase() *gorm.DB {
	db, err := gorm.Open("sqlite3", "database.db")
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&User{})
	db.AutoMigrate(&Exam{})
	db.AutoMigrate(&UserAllowedToExam{})
	db.AutoMigrate(&Question{})
	db.AutoMigrate(&Result{})
	db.AutoMigrate(&ExamParticipation{})

	return db
}
