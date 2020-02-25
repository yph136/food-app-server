package interfaces

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"food-app/application"
	"food-app/domain/entity"
	"food-app/utils/auth"
	"food-app/utils/fileupload"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var fakeT auth.TokenInterface = &fakeToken{}
var fakeA auth.AuthInterface = &fakeAuth{}
var fakeFood application.FoodAppInterface = &fakeFoodApp{}
var fakeUser application.UserAppInterface = &fakeUserApp{}

//var fakeFood repository.FoodRepository = &fakeFoodRepo{} //this is where the real implementation is swap with our fake implementation

//application.UserApp = &fakeUserApp{}


//application.FoodApp = &fakeFoodApp{} //make it possible to change real method with fake

var food = Food{}


//IF YOU HAVE TIME, YOU CAN TEST ALL FAILURE CASES TO IMPROVE COVERAGE

func Test_SaveFood_Invalid_Data(t *testing.T) {

	//var _ auth.AuthInterface = &fakeAuth{}


	//Mock extracting metadata
	tokenMetadata = func(r *http.Request) (*auth.AccessDetails, error){
		return &auth.AccessDetails{
			TokenUuid: "0237817a-1546-4ca3-96a4-17621c237f6b",
			UserId:    1,
		}, nil
	}
	//Mocking the fetching of token metadata from redis
	fetchAuth =  func(uuid string) (uint64, error){
		return 1, nil
	}
	samples := []struct {
		inputJSON  string
		statusCode int
	}{
		{
			//when the title is empty
			inputJSON:  `{"title": "", "description": "the desc"}`,
			statusCode: 422,
		},
		{
			//the description is empty
			inputJSON:  `{"title": "the title", "description": ""}`,
			statusCode: 422,
		},
		{
			//both the title and the description are empty
			inputJSON:  `{"title": "", "description": ""}`,
			statusCode: 422,
		},
		{
			//When invalid data is passed, e.g, instead of an integer, a string is passed
			inputJSON:  `{"title": 12344, "description": "the desc"}`,
			statusCode: 422,
		},
		{
			//When invalid data is passed, e.g, instead of an integer, a string is passed
			inputJSON:  `{"title": "hello title", "description": 3242342}`,
			statusCode: 422,
		},
	}

	for _, v := range samples {
		//use a valid token that has not expired. This token was created to live forever, just for test purposes with the user id of 1. This is so that it can always be used to run tests
		token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NfdXVpZCI6IjgyYTM3YWE5LTI4MGMtNDQ2OC04M2RmLTZiOGYyMDIzODdkMyIsImF1dGhvcml6ZWQiOnRydWUsInVzZXJfaWQiOjF9.ESelxq-UHormgXUwRNe4_Elz2i__9EKwCXPsNCyKV5o"
		tokenString := fmt.Sprintf("Bearer %v", token)

		r := gin.Default()
		r.POST("/food", food.SaveFood)
		req, err := http.NewRequest(http.MethodPost, "/food", bytes.NewBufferString(v.inputJSON))
		if err != nil {
			t.Errorf("this is the error: %v\n", err)
		}
		req.Header.Set("Authorization", tokenString)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		validationErr := make(map[string]string)

		err = json.Unmarshal(rr.Body.Bytes(), &validationErr)
		if err != nil {
			t.Errorf("error unmarshalling error %s\n", err)
		}
		assert.Equal(t, rr.Code, v.statusCode)

		if validationErr["title_required"] != "" {
			assert.Equal(t, validationErr["title_required"], "title is required")
		}
		if validationErr["description_required"] != "" {
			assert.Equal(t, validationErr["description_required"], "description is required")
		}
		if validationErr["title_required"] != "" && validationErr["description_required"] != "" {
			assert.Equal(t, validationErr["title_required"], "title is required")
			assert.Equal(t, validationErr["description_required"], "description is required")
		}
		if validationErr["invalid_json"] != "" {
			assert.Equal(t, validationErr["invalid_json"], "invalid json")
		}
	}
}


func TestSaverFood_Success(t *testing.T) {
	//application.FoodApp = &fakeFoodApp{} //make it possible to change real method with fake
	//auth.Token = &fakeToken{}
	//auth.Auth = &fakeAuth{}
	//application.UserApp = &fakeUserApp{}
	fileupload.Uploader = &fakeUploader{}

	//Mock extracting metadata
	tokenMetadata = func(r *http.Request) (*auth.AccessDetails, error){
		return &auth.AccessDetails{
			TokenUuid: "0237817a-1546-4ca3-96a4-17621c237f6b",
			UserId:    1,
		}, nil
	}
	//Mocking the fetching of token metadata from redis
	fetchAuth =  func(uuid string) (uint64, error){
		return 1, nil
	}
	getUserApp = func(uint64) (*entity.User, error) {
		//remember we are running sensitive info such as email and password
		return &entity.User{
			ID:        1,
			FirstName: "victor",
			LastName:  "steven",
		}, nil
	}
	//Mocking file upload to DigitalOcean
	uploadFile = func(file *multipart.FileHeader) (string, error) {
		return "dbdbf-dhbfh-bfy34-34jh-fd.jpg", nil //this is fabricated
	}
	//Mocking The Food return from db
	saveFoodApp = func(*entity.Food) (*entity.Food, map[string]string) {
		return &entity.Food{
			ID:        1,
			UserID:    1,
			Title: "Food title",
			Description:  "Food description",
			FoodImage: "dbdbf-dhbfh-bfy34-34jh-fd.jpg",

		}, nil
	}
	f :=  "./../utils/test_images/amala.jpg" //this is where the image is located
	file, err := os.Open(f)
	if err != nil {
		t.Errorf("Cannot open file: %s\n", err)
	}
	defer file.Close()

	//Create a buffer to store our request body as bytes
	var requestBody bytes.Buffer

	//Create a multipart writer
	multipartWriter := multipart.NewWriter(&requestBody)

	//Initialize the file field
	fileWriter, err := multipartWriter.CreateFormFile("food_image", "amala.jpg")
	if err != nil {
		t.Errorf("Cannot write file: %s\n", err)
	}
	//Copy the actual content to the file field's writer, though this is not needed, since files are sent to DigitalOcean
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		t.Errorf("Cannot copy file: %s\n", err)
	}
	//Add the title and the description fields
	fileWriter, err = multipartWriter.CreateFormField("title")
	if err != nil {
		t.Errorf("Cannot write title: %s\n", err)
	}
	_, err = fileWriter.Write([]byte("Food title"))
	if err != nil {
		t.Errorf("Cannot write title value: %s\n", err)
	}
	fileWriter, err = multipartWriter.CreateFormField("description")
	if err != nil {
		t.Errorf("Cannot write description: %s\n", err)
	}
	_, err = fileWriter.Write([]byte("Food description"))
	if err != nil {
		t.Errorf("Cannot write description value: %s\n", err)
	}
	//Close the multipart writer so it writes the ending boundary
	multipartWriter.Close()

	//use a valid token that has not expired. This token was created to live forever, just for test purposes with the user id of 1. This is so that it can always be used to run tests
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NfdXVpZCI6IjgyYTM3YWE5LTI4MGMtNDQ2OC04M2RmLTZiOGYyMDIzODdkMyIsImF1dGhvcml6ZWQiOnRydWUsInVzZXJfaWQiOjF9.ESelxq-UHormgXUwRNe4_Elz2i__9EKwCXPsNCyKV5o"

	tokenString := fmt.Sprintf("Bearer %v", token)

	req, err := http.NewRequest(http.MethodPost, "/food", &requestBody)
	if err != nil {
		t.Errorf("this is the error: %v\n", err)
	}
	r := gin.Default()
	r.POST("/food", food.SaveFood)
	req.Header.Set("Authorization", tokenString)
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType()) //this is important
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var food = entity.Food{}
	err = json.Unmarshal(rr.Body.Bytes(), &food)
	if err != nil {
		t.Errorf("cannot unmarshal response: %v\n", err)
	}
	assert.Equal(t, rr.Code, 201)
	assert.EqualValues(t, food.ID, 1)
	assert.EqualValues(t, food.UserID, 1)
	assert.EqualValues(t, food.Title, "Food title")
	assert.EqualValues(t, food.Description, "Food description")
	assert.EqualValues(t, food.FoodImage, "dbdbf-dhbfh-bfy34-34jh-fd.jpg")
}

//When wrong token is provided
func TestSaverFood_Unauthorized(t *testing.T) {
	//application.FoodApp = &fakeFoodApp{} //make it possible to change real method with fake
	//auth.Token = &fakeToken{}
	//auth.Auth = &fakeAuth{}
	//application.UserApp = &fakeUserApp{}
	fileupload.Uploader = &fakeUploader{}

	//Mock extracting metadata
	tokenMetadata = func(r *http.Request) (*auth.AccessDetails, error){
		return nil, errors.New("unauthorized")
	}

	f :=  "./../utils/test_images/amala.jpg" //this is where the image is located
	file, err := os.Open(f)
	if err != nil {
		t.Errorf("Cannot open file: %s\n", err)
	}
	defer file.Close()

	//Create a buffer to store our request body as bytes
	var requestBody bytes.Buffer

	//Create a multipart writer
	multipartWriter := multipart.NewWriter(&requestBody)

	//Initialize the file field
	fileWriter, err := multipartWriter.CreateFormFile("food_image", "amala.jpg")
	if err != nil {
		t.Errorf("Cannot write file: %s\n", err)
	}
	//Copy the actual content to the file field's writer, though this is not needed, since files are sent to DigitalOcean
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		t.Errorf("Cannot copy file: %s\n", err)
	}
	//Add the title and the description fields
	fileWriter, err = multipartWriter.CreateFormField("title")
	if err != nil {
		t.Errorf("Cannot write title: %s\n", err)
	}
	_, err = fileWriter.Write([]byte("Food title"))
	if err != nil {
		t.Errorf("Cannot write title value: %s\n", err)
	}
	fileWriter, err = multipartWriter.CreateFormField("description")
	if err != nil {
		t.Errorf("Cannot write description: %s\n", err)
	}
	_, err = fileWriter.Write([]byte("Food description"))
	if err != nil {
		t.Errorf("Cannot write description value: %s\n", err)
	}
	//Close the multipart writer so it writes the ending boundary
	multipartWriter.Close()

	token := "wrong-token-string"

	tokenString := fmt.Sprintf("Bearer %v", token)

	req, err := http.NewRequest(http.MethodPost, "/food", &requestBody)
	if err != nil {
		t.Errorf("this is the error: %v\n", err)
	}
	r := gin.Default()
	r.POST("/food", food.SaveFood)
	req.Header.Set("Authorization", tokenString)
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType()) //this is important
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var errResp = ""
	err = json.Unmarshal(rr.Body.Bytes(), &errResp)
	if err != nil {
		t.Errorf("cannot unmarshal response: %v\n", err)
	}
	assert.Equal(t, rr.Code, 401)
	assert.EqualValues(t, errResp, "unauthorized")
}

func TestGetAllFood_Success(t *testing.T) {
	//application.FoodApp = &fakeFoodApp{} //make it possible to change real method with fake

	//Return Food to check for, with our mock
	getAllFoodApp = func() ([]entity.Food, error) {
		return []entity.Food{
			{
				ID:        1,
				UserID:    1,
				Title: "Food title",
				Description:  "Food description",
				FoodImage: "dbdbf-dhbfh-bfy34-34jh-fd.jpg",
			},
			{
				ID:        2,
				UserID:    2,
				Title: "Food title second",
				Description:  "Food description second",
				FoodImage: "dbdbf-dhbfh-bfy34-34jh-fd-second.jpg",
			},
		}, nil
	}
	req, err := http.NewRequest(http.MethodGet, "/food", nil)
	if err != nil {
		t.Errorf("this is the error: %v\n", err)
	}
	r := gin.Default()
	r.GET("/food", food.GetAllFood)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var food []entity.Food
	err = json.Unmarshal(rr.Body.Bytes(), &food)
	if err != nil {
		t.Errorf("cannot unmarshal response: %v\n", err)
	}
	assert.Equal(t, rr.Code, 200)
	assert.EqualValues(t, len(food), 2)
}


func TestGetFoodAndCreator_Success(t *testing.T) {
	//application.FoodApp = &fakeFoodApp{} //make it possible to change real method with fake
	//application.UserApp = &fakeUserApp{}

	getUserApp = func(uint64) (*entity.User, error) {
		//remember we are running sensitive info such as email and password
		return &entity.User{
			ID:        1,
			FirstName: "victor",
			LastName:  "steven",
		}, nil
	}
	//Return Food to check for, with our mock
	getFoodApp = func(uint64) (*entity.Food, error) {
		return &entity.Food{
			ID:        1,
			UserID:    1,
			Title: "Food title",
			Description:  "Food description",
			FoodImage: "dbdbf-dhbfh-bfy34-34jh-fd.jpg",
		}, nil
	}
	foodID := strconv.Itoa(1)
	req, err := http.NewRequest(http.MethodGet, "/food/"+foodID, nil)
	if err != nil {
		t.Errorf("this is the error: %v\n", err)
	}
	r := gin.Default()
	r.GET("/food/:food_id", food.GetFoodAndCreator)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var foodAndCreator = make(map[string]interface{})
	err = json.Unmarshal(rr.Body.Bytes(), &foodAndCreator)
	if err != nil {
		t.Errorf("cannot unmarshal response: %v\n", err)
	}
	food := foodAndCreator["food"].(map[string]interface{})
	creator := foodAndCreator["creator"].(map[string]interface{})

	assert.Equal(t, rr.Code, 200)

	assert.EqualValues(t, food["title"], "Food title")
	assert.EqualValues(t, food["description"], "Food description")
	assert.EqualValues(t, food["food_image"], "dbdbf-dhbfh-bfy34-34jh-fd.jpg")

	assert.EqualValues(t, creator["first_name"], "victor")
	assert.EqualValues(t, creator["last_name"], "steven")
}


func TestUpdateFood_Success_With_File(t *testing.T) {
	//application.FoodApp = &fakeFoodApp{} //make it possible to change real method with fake
	//auth.Token = &fakeToken{}
	//auth.Auth = &fakeAuth{}
	//application.UserApp = &fakeUserApp{}
	fileupload.Uploader = &fakeUploader{}

	//Mock extracting metadata
	tokenMetadata = func(r *http.Request) (*auth.AccessDetails, error){
		return &auth.AccessDetails{
			TokenUuid: "0237817a-1546-4ca3-96a4-17621c237f6b",
			UserId:    1,
		}, nil
	}
	//Mocking the fetching of token metadata from redis
	fetchAuth =  func(uuid string) (uint64, error){
		return 1, nil
	}
	getUserApp = func(uint64) (*entity.User, error) {
		//remember we are running sensitive info such as email and password
		return &entity.User{
			ID:        1,
			FirstName: "victor",
			LastName:  "steven",
		}, nil
	}
	//Return Food to check for, with our mock
	getFoodApp = func(uint64) (*entity.Food, error) {
		return &entity.Food{
			ID:        1,
			UserID:    1,
			Title: "Food title",
			Description:  "Food description",
			FoodImage: "dbdbf-dhbfh-bfy34-34jh-fd.jpg",
		}, nil
	}
	//Mocking The Food return from db
	updateFoodApp = func(*entity.Food) (*entity.Food, map[string]string) {
		return &entity.Food{
			ID:        1,
			UserID:    1,
			Title: "Food title updated",
			Description:  "Food description updated",
			FoodImage: "dbdbf-dhbfh-bfy34-34jh-fd-updated.jpg",
		}, nil
	}

	//Mocking file upload to DigitalOcean
	uploadFile = func(file *multipart.FileHeader) (string, error) {
		return "dbdbf-dhbfh-bfy34-34jh-fd-updated.jpg", nil //this is fabricated
	}

	f :=  "./../utils/test_images/new_meal.jpeg" //this is where the image is located
	file, err := os.Open(f)
	if err != nil {
		t.Errorf("Cannot open file: %s\n", err)
	}
	defer file.Close()

	//Create a buffer to store our request body as bytes
	var requestBody bytes.Buffer

	//Create a multipart writer
	multipartWriter := multipart.NewWriter(&requestBody)

	//Initialize the file field
	fileWriter, err := multipartWriter.CreateFormFile("food_image", "new_meal.jpeg")
	if err != nil {
		t.Errorf("Cannot write file: %s\n", err)
	}
	//Copy the actual content to the file field's writer, though this is not needed, since files are sent to DigitalOcean
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		t.Errorf("Cannot copy file: %s\n", err)
	}
	//Add the title and the description fields
	fileWriter, err = multipartWriter.CreateFormField("title")
	if err != nil {
		t.Errorf("Cannot write title: %s\n", err)
	}
	_, err = fileWriter.Write([]byte("Food title updated"))
	if err != nil {
		t.Errorf("Cannot write title value: %s\n", err)
	}
	fileWriter, err = multipartWriter.CreateFormField("description")
	if err != nil {
		t.Errorf("Cannot write description: %s\n", err)
	}
	_, err = fileWriter.Write([]byte("Food description updated"))
	if err != nil {
		t.Errorf("Cannot write description value: %s\n", err)
	}
	//Close the multipart writer so it writes the ending boundary
	multipartWriter.Close()

	//use a valid token that has not expired. This token was created to live forever, just for test purposes with the user id of 1. This is so that it can always be used to run tests
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NfdXVpZCI6IjgyYTM3YWE5LTI4MGMtNDQ2OC04M2RmLTZiOGYyMDIzODdkMyIsImF1dGhvcml6ZWQiOnRydWUsInVzZXJfaWQiOjF9.ESelxq-UHormgXUwRNe4_Elz2i__9EKwCXPsNCyKV5o"

	tokenString := fmt.Sprintf("Bearer %v", token)

	foodID := strconv.Itoa(1)
	req, err := http.NewRequest(http.MethodPut, "/food/"+foodID, &requestBody)
	if err != nil {
		t.Errorf("this is the error: %v\n", err)
	}
	r := gin.Default()
	r.PUT("/food/:food_id", food.UpdateFood)
	req.Header.Set("Authorization", tokenString)
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType()) //this is important
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var food = entity.Food{}
	err = json.Unmarshal(rr.Body.Bytes(), &food)
	if err != nil {
		t.Errorf("cannot unmarshal response: %v\n", err)
	}
	assert.Equal(t, rr.Code, 200)
	assert.EqualValues(t, food.ID, 1)
	assert.EqualValues(t, food.UserID, 1)
	assert.EqualValues(t, food.Title, "Food title updated")
	assert.EqualValues(t, food.Description, "Food description updated")
	assert.EqualValues(t, food.FoodImage, "dbdbf-dhbfh-bfy34-34jh-fd-updated.jpg")
}


//This is where file is not updated. A user can choose not to update file, in that case, the old file will still be used
func TestUpdateFood_Success_Without_File(t *testing.T) {
	//application.FoodApp = &fakeFoodApp{} //make it possible to change real method with fake
	//auth.Token = &fakeToken{}
	//auth.Auth = &fakeAuth{}
	//application.UserApp = &fakeUserApp{}
	fileupload.Uploader = &fakeUploader{}

	//Mock extracting metadata
	tokenMetadata = func(r *http.Request) (*auth.AccessDetails, error){
		return &auth.AccessDetails{
			TokenUuid: "0237817a-1546-4ca3-96a4-17621c237f6b",
			UserId:    1,
		}, nil
	}
	//Mocking the fetching of token metadata from redis
	fetchAuth =  func(uuid string) (uint64, error){
		return 1, nil
	}
	getUserApp = func(uint64) (*entity.User, error) {
		//remember we are running sensitive info such as email and password
		return &entity.User{
			ID:        1,
			FirstName: "victor",
			LastName:  "steven",
		}, nil
	}
	//Return Food to check for, with our mock
	getFoodApp = func(uint64) (*entity.Food, error) {
		return &entity.Food{
			ID:        1,
			UserID:    1,
			Title: "Food title",
			Description:  "Food description",
			FoodImage: "dbdbf-dhbfh-bfy34-34jh-fd-old-file.jpg",
		}, nil
	}
	//Mocking The Food return from db
	updateFoodApp = func(*entity.Food) (*entity.Food, map[string]string) {
		return &entity.Food{
			ID:        1,
			UserID:    1,
			Title: "Food title updated",
			Description:  "Food description updated",
			FoodImage: "dbdbf-dhbfh-bfy34-34jh-fd-old-file.jpg",
		}, nil
	}

	//Mocking file upload to DigitalOcean
	uploadFile = func(file *multipart.FileHeader) (string, error) {
		return "dbdbf-dhbfh-bfy34-34jh-fd-old-file.jpg", nil //this is fabricated
	}

	//Create a buffer to store our request body as bytes
	var requestBody bytes.Buffer

	//Create a multipart writer
	multipartWriter := multipart.NewWriter(&requestBody)

	//Add the title and the description fields
	fileWriter, err := multipartWriter.CreateFormField("title")
	if err != nil {
		t.Errorf("Cannot write title: %s\n", err)
	}
	_, err = fileWriter.Write([]byte("Food title updated"))
	if err != nil {
		t.Errorf("Cannot write title value: %s\n", err)
	}
	fileWriter, err = multipartWriter.CreateFormField("description")
	if err != nil {
		t.Errorf("Cannot write description: %s\n", err)
	}
	_, err = fileWriter.Write([]byte("Food description updated"))
	if err != nil {
		t.Errorf("Cannot write description value: %s\n", err)
	}
	//Close the multipart writer so it writes the ending boundary
	multipartWriter.Close()

	//use a valid token that has not expired. This token was created to live forever, just for test purposes with the user id of 1. This is so that it can always be used to run tests
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NfdXVpZCI6IjgyYTM3YWE5LTI4MGMtNDQ2OC04M2RmLTZiOGYyMDIzODdkMyIsImF1dGhvcml6ZWQiOnRydWUsInVzZXJfaWQiOjF9.ESelxq-UHormgXUwRNe4_Elz2i__9EKwCXPsNCyKV5o"

	tokenString := fmt.Sprintf("Bearer %v", token)

	foodID := strconv.Itoa(1)
	req, err := http.NewRequest(http.MethodPut, "/food/"+foodID, &requestBody)
	if err != nil {
		t.Errorf("this is the error: %v\n", err)
	}
	r := gin.Default()
	r.PUT("/food/:food_id", food.UpdateFood)
	req.Header.Set("Authorization", tokenString)
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType()) //this is important
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var food = entity.Food{}
	err = json.Unmarshal(rr.Body.Bytes(), &food)
	if err != nil {
		t.Errorf("cannot unmarshal response: %v\n", err)
	}
	assert.Equal(t, rr.Code, 200)
	assert.EqualValues(t, food.ID, 1)
	assert.EqualValues(t, food.UserID, 1)
	assert.EqualValues(t, food.Title, "Food title updated")
	assert.EqualValues(t, food.Description, "Food description updated")
	assert.EqualValues(t, food.FoodImage, "dbdbf-dhbfh-bfy34-34jh-fd-old-file.jpg")
}


func TestUpdateFood_Invalid_Data(t *testing.T) {
	//auth.Token = &fakeToken{}
	//auth.Auth = &fakeAuth{}

	//Mock extracting metadata
	tokenMetadata = func(r *http.Request) (*auth.AccessDetails, error){
		return &auth.AccessDetails{
			TokenUuid: "0237817a-1546-4ca3-96a4-17621c237f6b",
			UserId:    1,
		}, nil
	}
	//Mocking the fetching of token metadata from redis
	fetchAuth =  func(uuid string) (uint64, error){
		return 1, nil
	}

	samples := []struct {
		inputJSON  string
		statusCode int
	}{
		{
			//when the title is empty
			inputJSON:  `{"title": "", "description": "the desc"}`,
			statusCode: 422,
		},
		{
			//the description is empty
			inputJSON:  `{"title": "the title", "description": ""}`,
			statusCode: 422,
		},
		{
			//both the title and the description are empty
			inputJSON:  `{"title": "", "description": ""}`,
			statusCode: 422,
		},
		{
			//When invalid data is passed, e.g, instead of an integer, a string is passed
			inputJSON:  `{"title": 12344, "description": "the desc"}`,
			statusCode: 422,
		},
		{
			//When invalid data is passed, e.g, instead of an integer, a string is passed
			inputJSON:  `{"title": "hello sir", "description": 3242342}`,
			statusCode: 422,
		},
	}

	for _, v := range samples {

		//use a valid token that has not expired. This token was created to live forever, just for test purposes with the user id of 1. This is so that it can always be used to run tests
		token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NfdXVpZCI6IjgyYTM3YWE5LTI4MGMtNDQ2OC04M2RmLTZiOGYyMDIzODdkMyIsImF1dGhvcml6ZWQiOnRydWUsInVzZXJfaWQiOjF9.ESelxq-UHormgXUwRNe4_Elz2i__9EKwCXPsNCyKV5o"
		tokenString := fmt.Sprintf("Bearer %v", token)

		foodID := strconv.Itoa(1)

		r := gin.Default()
		r.POST("/food/:food_id", food.UpdateFood)
		req, err := http.NewRequest(http.MethodPost, "/food/"+foodID, bytes.NewBufferString(v.inputJSON))
		if err != nil {
			t.Errorf("this is the error: %v\n", err)
		}
		req.Header.Set("Authorization", tokenString)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		validationErr := make(map[string]string)

		err = json.Unmarshal(rr.Body.Bytes(), &validationErr)
		if err != nil {
			t.Errorf("error unmarshalling error %s\n", err)
		}
		assert.Equal(t, rr.Code, v.statusCode)


		if validationErr["title_required"] != "" {
			assert.Equal(t, validationErr["title_required"], "title is required")
		}
		if validationErr["description_required"] != "" {
			assert.Equal(t, validationErr["description_required"], "description is required")
		}
		if validationErr["title_required"] != "" && validationErr["description_required"] != "" {
			assert.Equal(t, validationErr["title_required"], "title is required")
			assert.Equal(t, validationErr["description_required"], "description is required")
		}
		if validationErr["invalid_json"] != "" {
			assert.Equal(t, validationErr["invalid_json"], "invalid json")
		}
	}
}

func TestDeleteFood_Success(t *testing.T) {
	//application.FoodApp = &fakeFoodApp{} //make it possible to change real method with fake
	//auth.Token = &fakeToken{}
	//auth.Auth = &fakeAuth{}
	//application.UserApp = &fakeUserApp{}

	//Mock extracting metadata
	tokenMetadata = func(r *http.Request) (*auth.AccessDetails, error){
		return &auth.AccessDetails{
			TokenUuid: "0237817a-1546-4ca3-96a4-17621c237f6b",
			UserId:    1,
		}, nil
	}
	//Mocking the fetching of token metadata from redis
	fetchAuth =  func(uuid string) (uint64, error){
		return 1, nil
	}
	//Return Food to check for, with our mock
	getFoodApp = func(uint64) (*entity.Food, error) {
		return &entity.Food{
			ID:        1,
			UserID:    1,
			Title: "Food title",
			Description:  "Food description",
			FoodImage: "dbdbf-dhbfh-bfy34-34jh-fd-old-file.jpg",
		}, nil
	}
	getUserApp = func(uint64) (*entity.User, error) {
		//remember we are running sensitive info such as email and password
		return &entity.User{
			ID:        1,
			FirstName: "victor",
			LastName:  "steven",
		}, nil
	}
	//The deleted food mock:
	deleteFoodApp = func(uint64) error {
		return nil
	}
	//use a valid token that has not expired. This token was created to live forever, just for test purposes with the user id of 1. This is so that it can always be used to run tests
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NfdXVpZCI6IjgyYTM3YWE5LTI4MGMtNDQ2OC04M2RmLTZiOGYyMDIzODdkMyIsImF1dGhvcml6ZWQiOnRydWUsInVzZXJfaWQiOjF9.ESelxq-UHormgXUwRNe4_Elz2i__9EKwCXPsNCyKV5o"

	tokenString := fmt.Sprintf("Bearer %v", token)

	foodId := strconv.Itoa(1)
	req, err := http.NewRequest(http.MethodDelete, "/food/"+foodId, nil)
	if err != nil {
		t.Errorf("this is the error: %v\n", err)
	}
	r := gin.Default()
	r.DELETE("/food/:food_id", food.DeleteFood)
	req.Header.Set("Authorization", tokenString)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	response := ""

	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("cannot unmarshal response: %v\n", err)
	}
	assert.Equal(t, rr.Code, 200)
	assert.EqualValues(t, response, "food deleted")
}


