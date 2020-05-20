package examwebportal

import (
	"fmt"
	"net/http"
	"os"
)

func SendKavenegarOneToken(template string, phoneNo string, token string) {
	apiKey := getAPIKeyFromENV()

	req := "https://api.kavenegar.com/v1/" + apiKey + "/verify/lookup.json?receptor=" + phoneNo + "&token=" + token + "&template=" + template

	fmt.Println(req)
	res, _ := http.Get(req)
	fmt.Println(res)
}

func getAPIKeyFromENV() string {
	key := os.Getenv("KAVENEGAR_API_KEY")
	fmt.Println(key)
	return key
}
