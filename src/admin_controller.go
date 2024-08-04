package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
    "os"
    "io"
    "path/filepath"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func HandleAdminEndpoints(){
    http.HandleFunc("/admin", handleAdmin)
    http.HandleFunc("/authenticate", handleAuthentication)
    http.HandleFunc("/register", handleRegistration)
    http.HandleFunc("/parse-md", handleParseMarkdown)
    http.HandleFunc("/create-post", handlePostCreation)
    http.HandleFunc("/edit-post", handlePostEdit)
    http.HandleFunc("/delete-post", handlePostDeletion)
    http.HandleFunc("/upload", handleFileUpload)
    http.HandleFunc("/create-portfolio", handlePortfolioEntry)
    http.HandleFunc("/posts-management", handlePostManagement)
    http.HandleFunc("/portfolio-management", handlePortfolioManagement)
}


func handleAdmin(w http.ResponseWriter, r *http.Request){
    if !isAuthenticated(r){
        showLoginForm(w, r)
    } else{
        showDashboard(w, r, Post{})
    }
}

func isAuthenticated(r *http.Request) bool{
    cookie, err := r.Cookie("session_token")
    if err != nil {
        return false
    }

    collection := Client.Database("blog").Collection("sessions")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    var session struct {
        UserId  primitive.ObjectID `bson:"user_id"`
        Token   string             `bson:"token"`
        Expires time.Time          `bson:"expires"`
    }

    err = collection.FindOne(ctx, bson.M{"token": cookie.Value}).Decode(&session)
    if err != nil {
        return false
    }

    return session.Expires.After(time.Now())
}

func showLoginForm(w http.ResponseWriter, r *http.Request) {
    tmpl, err := ParseTemplates()
    if err != nil {
        log.Printf("Error loading templates: %v\n", err)
        http.Error(w, "Error loading templates.", http.StatusInternalServerError)
        return
    }
    content, err := RenderTemplate(tmpl, "login", nil)
    if err != nil {
        log.Printf("Error rendering login template: %v\n", err)
        http.Error(w, "Error rendering template.", http.StatusInternalServerError)
        return
    }
    version := uuid.New()
    data := PageData{
        Content: template.HTML(content),
        Version: version.String(),
    }
    if err := tmpl.ExecuteTemplate(w, "admin", data); err != nil {
        log.Println(err)
        http.Error(w, "Error executing template.", http.StatusInternalServerError)
    }
}

type Editable struct{
    Post Post
    MarkDown template.HTML
}

type DashBoard struct{
    Editable Editable
    Posts []Post
}

func showDashboard(w http.ResponseWriter, r *http.Request, p Post) {
    tmpl, err := ParseTemplates()
    if err != nil {
        log.Printf("Error loading templates: %v\n", err)
        http.Error(w, "Error loading templates.", http.StatusInternalServerError)
        return
    }

    posts, err := GetPosts()
    if err != nil{
        http.Error(w, "Couldn't fetch posts", http.StatusInternalServerError)
    }

    dashBoard := DashBoard{
        Posts: posts,
        Editable: Editable{
            Post: p,
            MarkDown: template.HTML(MdToHtml([]byte(p.Body))),
        },
    }

    content, err := RenderTemplate(tmpl, "dashboard", dashBoard)
    if err != nil {
        log.Printf("Error rendering dashboard template: %v\n", err)
        http.Error(w, "Error rendering template.", http.StatusInternalServerError)
        return
    }
    version := uuid.New()
    data := PageData{
        Content: template.HTML(content),
        Version: version.String(),
    }
    if err := tmpl.ExecuteTemplate(w, "admin", data); err != nil {
        log.Println(err)
        http.Error(w, "Error executing template.", http.StatusInternalServerError)
    }
}

func handleAuthentication(w http.ResponseWriter, r *http.Request){
    if r.Method != http.MethodPost {
        http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
        return
    }

    username := r.FormValue("username")
    password := r.FormValue("password")

    user, err := AuthenticateUser(username, password)
    if err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    sessionToken := uuid.New().String()
    expiresAt := time.Now().Add(24 * time.Hour)

    http.SetCookie(w, &http.Cookie{
        Name:    "session_token",
        Value:   sessionToken,
        Expires: expiresAt,
    })

    sessionErr := storeSession(user.Id, sessionToken, expiresAt)

    if sessionErr != nil{
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func storeSession(userId primitive.ObjectID, token string, expiresAt time.Time) error {
    collection := Client.Database("blog").Collection("sessions")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    _, err := collection.InsertOne(ctx, bson.M{
        "user_id": userId,
        "token":   token,
        "expires": expiresAt,
    })
    return err
}

func handleRegistration(w http.ResponseWriter, r *http.Request){
    if r.Method == http.MethodGet{
        tmpl, err := ParseTemplates()
        if err != nil {
            log.Printf("Error loading templates: %v\n", err)
            http.Error(w, "Error loading templates.", http.StatusInternalServerError)
            return
        }
        content, err := RenderTemplate(tmpl, "register", nil)
        if err != nil {
            log.Printf("Error rendering dashboard template: %v\n", err)
            http.Error(w, "Error rendering template.", http.StatusInternalServerError)
            return
        }
        version := uuid.New()
        data := PageData{
            Content: template.HTML(content),
            Version: version.String(),
        }
        if err := tmpl.ExecuteTemplate(w, "register", data); err != nil {
            log.Println(err)
            http.Error(w, "Error executing template.", http.StatusInternalServerError)
        }

        return
    }
    if r.Method == http.MethodPost{
        username := r.FormValue("username")
        password := r.FormValue("password")

        if len(username) == 0 || len(password) == 0{
            http.Error(w, "Required fields not filled", http.StatusBadRequest)
        }

        err := RegisterUser(username, password)
        if err != nil{
            log.Fatal(err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }

        user, err := AuthenticateUser(username, password)
        if err != nil{
            log.Fatal(err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }

        sessionToken := uuid.New().String()
        expiresAt := time.Now().Add(24 * time.Hour)

        http.SetCookie(w, &http.Cookie{
            Name:    "session_token",
            Value:   sessionToken,
            Expires: expiresAt,
        })

        sessionErr := storeSession(user.Id, sessionToken, expiresAt)

        if sessionErr != nil {
            log.Printf("Error storing session: %v\n", sessionErr)
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }

        http.Redirect(w, r, "/admin", http.StatusSeeOther)
        return
    }

    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleParseMarkdown(w http.ResponseWriter, r *http.Request){
    if err := r.ParseForm(); err != nil{
        http.Error(w, "Failed to parse request", http.StatusBadRequest)
        return
    }

    text := r.FormValue("post-text")
    html := MdToHtml([]byte(text))
   
    w.Header().Set("Content-Type", "text/html")
    if _, err := w.Write(html); err != nil{
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
}

func handlePostCreation(w http.ResponseWriter, r *http.Request){
    if r.Method != http.MethodPost{
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    idStr := r.FormValue("post-id")
    title := r.FormValue("title")
    body := r.FormValue("post-text")
    synopsys := r.FormValue("synopsys")
    coverImage := r.FormValue("cover-image")

    if idStr != primitive.NilObjectID.Hex(){
        id, err := primitive.ObjectIDFromHex(idStr)
        if err != nil{
            http.Error(w, "Invalid Id", http.StatusInternalServerError)
            return
        }

        post, err := UpdatePost(id, title, body, synopsys, coverImage)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        showDashboard(w, r, post)
    } else { 
        post, err := CreatePost(title, body, synopsys, coverImage)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        showDashboard(w, r, post)
    }
}

func handlePostEdit(w http.ResponseWriter, r *http.Request){
    if r.Method != http.MethodGet{
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    if query := r.URL.Query(); !query.Has("id"){
        http.Error(w, "Post id is required", http.StatusBadRequest)
        return
    }

    id := r.URL.Query().Get("id")

    post, err := GetPost(id); 
    if err != nil{
        fmt.Println(err)
        http.Error(w, "Error fetching post", http.StatusInternalServerError)
        return
    }

    tmpl, err := ParseTemplates()
    if err != nil {
        log.Printf("Error loading templates: %v\n", err)
        http.Error(w, "Error loading templates.", http.StatusInternalServerError)
        return
    }

    editable := Editable{
        Post: post,
        MarkDown: template.HTML(MdToHtml([]byte(post.Body))),
    } 

	if err := tmpl.ExecuteTemplate(w, "post_form", editable); err != nil {
		log.Println(err)
		http.Error(w, "Error executing template.", http.StatusInternalServerError)
	}
}

func handlePostDeletion(w http.ResponseWriter, r *http.Request){
    if r.Method != http.MethodPost{
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    if query := r.URL.Query(); !query.Has("id"){
        http.Error(w, "Post id is required", http.StatusBadRequest)
        return
    }

    id := r.URL.Query().Get("id")

    if err := DeletePost(id); err != nil{
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    if _, err := w.Write([]byte("")); err != nil{
        http.Error(w, "Error creating response", http.StatusInternalServerError)
        return
    }

    fmt.Println(w)
}

func handleFileUpload(w http.ResponseWriter, r *http.Request){
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    err := r.ParseMultipartForm(10 << 20)
    if err != nil {
        http.Error(w, "Unable to parse form", http.StatusBadRequest)
        return
    }

    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Error retrieving the file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    ext := filepath.Ext(header.Filename)
    filename := uuid.New().String() + ext

    uploadDir := "./uploads"
    if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
        http.Error(w, "Unable to create upload directory", http.StatusInternalServerError)
        return
    }

    dst, err := os.Create(filepath.Join(uploadDir, filename))
    if err != nil {
        http.Error(w, "Unable to create the file", http.StatusInternalServerError)
        return
    }
    defer dst.Close()

    if _, err := io.Copy(dst, file); err != nil {
        http.Error(w, "Unable to write file", http.StatusInternalServerError)
        return
    }

    tmpl, err := ParseTemplates()
    if err != nil {
        log.Printf("Error loading templates: %v\n", err)
        http.Error(w, "Error loading templates.", http.StatusInternalServerError)
        return
    }

	if err := tmpl.ExecuteTemplate(w, "cover-image-field", Post{ CoverImage: filename}); err != nil {
		log.Println(err)
		http.Error(w, "Error executing template.", http.StatusInternalServerError)
	}
}

func handlePortfolioEntry(w http.ResponseWriter, r *http.Request){
    
}

func handlePostManagement(w http.ResponseWriter, r *http.Request){
    tmpl, err := ParseTemplates()
    if err != nil {
        log.Printf("Error loading templates: %v\n", err)
        http.Error(w, "Error loading templates.", http.StatusInternalServerError)
        return
    }

    posts, err := GetPosts()
    if err != nil{
        http.Error(w, "Couldn't fetch posts", http.StatusInternalServerError)
    }

    p := Post{}

    dashBoard := DashBoard{
        Posts: posts,
        Editable: Editable{
            Post: p,
            MarkDown: template.HTML(MdToHtml([]byte(p.Body))),
        },
    }

    if err := tmpl.ExecuteTemplate(w, "dashboard", dashBoard); err != nil {
        log.Println(err)
        http.Error(w, "Error executing template.", http.StatusInternalServerError)
    }
}

type PortfolioManagement struct{
    
}

func handlePortfolioManagement(w http.ResponseWriter, r *http.Request){
    tmpl, err := ParseTemplates()
    if err != nil {
        log.Printf("Error loading templates: %v\n", err)
        http.Error(w, "Error loading templates.", http.StatusInternalServerError)
        return
    }

    posts, err := GetPosts()
    if err != nil{
        http.Error(w, "Couldn't fetch posts", http.StatusInternalServerError)
    }

    p := Post{}

    dashBoard := DashBoard{
        Posts: posts,
        Editable: Editable{
            Post: p,
            MarkDown: template.HTML(MdToHtml([]byte(p.Body))),
        },
    }

    if err := tmpl.ExecuteTemplate(w, "portfolio-management", dashBoard); err != nil {
        log.Println(err)
        http.Error(w, "Error executing template.", http.StatusInternalServerError)
    }
}
