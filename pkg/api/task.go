package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Pipelines-Marketplace/backend/pkg/authentication"
	"github.com/Pipelines-Marketplace/backend/pkg/models"
	"github.com/Pipelines-Marketplace/backend/pkg/polling"
	"github.com/Pipelines-Marketplace/backend/pkg/upload"
	"github.com/Pipelines-Marketplace/backend/pkg/utility"
	"github.com/gorilla/mux"
)

// GetAllResources writes json encoded resources to ResponseWriter
func GetAllResources(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.GetAllResources())
}

// GetResourceByID writes json encoded resources to ResponseWriter
func GetResourceByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resourceID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"status": false, "message": "Invalid User ID"})
	}
	json.NewEncoder(w).Encode(models.GetResourceByID(resourceID))
}

// GetTaskFiles returns a compressed zip with task files
func GetTaskFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/zip")
	GetCompressedFiles(mux.Vars(r)["name"])
	// Serve the created zip file
	http.ServeFile(w, r, "finalZipFile.zip")
}

// GetAllTags writes json encoded list of tags to Responsewriter
func GetAllTags(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.GetAllTags())
}

// GetAllFilteredResourcesByTag writes json encoded list of filtered tasks to Responsewriter
func GetAllFilteredResourcesByTag(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.GetAllResourcesWithGivenTags(strings.Split(r.FormValue("tags"), "|")))
}

// GetResourceYAMLFile returns a compressed zip with task files
func GetResourceYAMLFile(w http.ResponseWriter, r *http.Request) {
	resourceID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		log.Println(err)
	}
	githubDetails := models.GetResourceGithubDetails(resourceID)
	desc, err := polling.GetFileContent(utility.Ctx, utility.Client, githubDetails.Owner, githubDetails.RepositoryName, githubDetails.Path, nil)
	if err != nil {
		log.Println(err)
	}
	content, err := desc.GetContent()
	if err != nil {
		log.Println(err)
	}
	w.Write([]byte(content))
}

// GetResourceReadmeFile returns a compressed zip with task files
func GetResourceReadmeFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/file")
	taskID := mux.Vars(r)["id"]
	readmeExists := models.DoesREADMEExist(taskID)
	if readmeExists {
		http.ServeFile(w, r, "readme/"+taskID+".md")
	}
	json.NewEncoder(w).Encode("noreadme")
}

// DownloadFile returns a requested YAML file
func DownloadFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/file")
	taskID := mux.Vars(r)["id"]
	models.IncrementDownloads(taskID)
	http.ServeFile(w, r, "tekton/"+taskID+".yaml")
}

// UpdateRating will add a new rating
func UpdateRating(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ratingRequestBody := AddRatingsRequest{}
	err := json.NewDecoder(r.Body).Decode(&ratingRequestBody)
	if err != nil {
		log.Println(err)
	}
	json.NewEncoder(w).Encode(models.UpdateRating(ratingRequestBody.UserID, ratingRequestBody.ResourceID, ratingRequestBody.Stars, ratingRequestBody.PrevStars))
}

// GetRatingDetails returns rating details of a task
func GetRatingDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resourceID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		log.Println(err)
	}
	json.NewEncoder(w).Encode(models.GetRatingDetialsByResourceID(resourceID))
}

// AddRating add's a new rating
func AddRating(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ratingRequestBody := AddRatingsRequest{}
	err := json.NewDecoder(r.Body).Decode(&ratingRequestBody)
	if err != nil {
		log.Println(err)
	}
	json.NewEncoder(w).Encode(models.AddRating(ratingRequestBody.UserID, ratingRequestBody.ResourceID, ratingRequestBody.Stars, ratingRequestBody.PrevStars))
}

// Upload a new task/pipeline
func Upload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	uploadRequestBody := upload.NewUploadRequestObject{}
	err := json.NewDecoder(r.Body).Decode(&uploadRequestBody)
	if err != nil {
		log.Println(err)
	}
	json.NewEncoder(w).Encode(upload.NewUpload(uploadRequestBody.Name, uploadRequestBody.Description, uploadRequestBody.Type, uploadRequestBody.Tags, uploadRequestBody.Github, uploadRequestBody.UserID))
}

// GetPrevStars will return the previous rating
func GetPrevStars(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	previousStarRequestBody := models.PrevStarRequest{}
	err := json.NewDecoder(r.Body).Decode(&previousStarRequestBody)
	if err != nil {
		log.Println(err)
	}
	json.NewEncoder(w).Encode(models.GetUserRating(previousStarRequestBody.UserID, previousStarRequestBody.ResourceID))

}

// OAuthAccessResponse represents access_token
type OAuthAccessResponse struct {
	AccessToken string `json:"access_token"`
}

// Code will
type Code struct {
	Token string `json:"token"`
}

// GithubAuth handles OAuth by Github
func GithubAuth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	httpClient := http.Client{}
	var clientID string
	var clientSecret string
	clientID = os.Getenv("CLIENT_ID")
	clientSecret = os.Getenv("CLIENT_SECRET")
	token := Code{}
	err := json.NewDecoder(r.Body).Decode(&token)
	if err != nil {
		log.Println(err)
	}
	log.Println("Code", token.Token)
	var code string
	code = token.Token
	reqURL := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s", clientID, clientSecret, code)
	log.Println(reqURL)
	req, err := http.NewRequest(http.MethodPost, reqURL, nil)
	if err != nil {
		println(os.Stdout, "could not create HTTP request: %v", err)
	}
	req.Header.Set("accept", "application/json")

	// Send out the HTTP request
	res, err := httpClient.Do(req)
	if err != nil {
		println(os.Stdout, "could not send HTTP request: %v", err)
	}

	// Parse the request body into the `OAuthAccessResponse` struct
	var t OAuthAccessResponse
	if err := json.NewDecoder(res.Body).Decode(&t); err != nil {
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
	}
	log.Println("Access Token", t.AccessToken)
	username, id := getUserDetails(t.AccessToken)
	log.Println(username, id)
	authToken, err := authentication.GenerateJWT(int(id))
	if err != nil {
		log.Println(err)
	}
	// Add user if doesn't exist
	sqlStatement := `SELECT EXISTS(SELECT 1 FROM USER_CREDENTIAL WHERE ID=$1)`
	var exists bool
	err = models.DB.QueryRow(sqlStatement, id).Scan(&exists)
	if err != nil {
		log.Println(err)
	}
	log.Println(exists)
	if !exists {
		sqlStatement := `INSERT INTO USER_CREDENTIAL(ID,USERNAME,FIRST_NAME) VALUES($1,$2,$3)`
		_, err := models.DB.Exec(sqlStatement, id, "github", "github")
		if err != nil {
			log.Println(err)
		}
	}
	// Update token here
	sqlStatement = `UPDATE USER_CREDENTIAL SET TOKEN=$2 WHERE ID=$1`
	_, err = models.DB.Exec(sqlStatement, id, t.AccessToken)
	if err != nil {
		log.Println(err)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"token": authToken, "user_id": int(id)})
}

func getUserDetails(accessToken string) (string, int) {
	httpClient := http.Client{}
	reqURL := fmt.Sprintf("https://api.github.com/user")
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	req.Header.Set("Authorization", "token "+accessToken)
	if err != nil {
		log.Println(err)
	}
	req.Header.Set("Access-Control-Allow-Origin", "*")
	req.Header.Set("accept", "application/json")

	// Send out the HTTP request
	res, err := httpClient.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	log.Println(string(body))
	var userData map[string]interface{}
	if err := json.Unmarshal(body, &userData); err != nil {
		log.Println(err)
	}
	username := userData["login"].(string)
	id := userData["id"].(float64)
	log.Println(id)
	return string(username), int(id)
}

// GetAllTasksByUserHandler will return all tasks uploaded by user
func GetAllTasksByUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"status": false, "message": "Invalid User ID"})
	}
	json.NewEncoder(w).Encode(models.GetAllTasksByUser(userID))

}
