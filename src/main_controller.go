package main

import (
    "html/template"
    "log"
    "net/http"
    "fmt"
    "regexp"
    "github.com/google/uuid"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

func HandleEndpoints(){ 
    fs := http.FileServer(http.Dir("./dist"))
    http.Handle("/dist/", http.StripPrefix("/dist/", fs))
    http.HandleFunc("/", handleIndex)
    http.HandleFunc("/contact", handleContact)
    http.HandleFunc("/home", handleHome)
    http.HandleFunc("/blog", handleBlog)
    http.HandleFunc("/like", handleLikeIncrement)
    http.HandleFunc("/post", handleReadPost)
    http.HandleFunc("/search-posts", handleSearchPosts)
}

func handleIndex(w http.ResponseWriter, r *http.Request){
    tmpl, err := ParseTemplates()
    if err != nil {
        log.Printf("Error loading templates: %v\n", err)
        http.Error(w, "Error loading templates.", http.StatusInternalServerError)
        return
    }

    content, err := RenderTemplate(tmpl, "home", nil)
    if err != nil {
        log.Printf("Error rendering home template: %v\n", err)
        http.Error(w, "Error rendering template.", http.StatusInternalServerError)
        return
    }

    version := uuid.New()
    data := PageData{
        Content: template.HTML(content),
        Version: version.String(),
    }

    if err := tmpl.ExecuteTemplate(w, "index", data); err != nil {
        log.Println(err)
        http.Error(w, "Error executing template.", http.StatusInternalServerError)
    }
}

func handleContact(w http.ResponseWriter, r *http.Request){
    tmpl, err := ParseTemplates()
    if err != nil {
        log.Printf("Error loading templates: %v\n", err)
        http.Error(w, "Error loading templates.", http.StatusInternalServerError)
        return
    }

    content, err := RenderTemplate(tmpl, "contact", nil)
    if err != nil {
        log.Printf("Error rendering contact template: %v\n", err)
        http.Error(w, "Error rendering template.", http.StatusInternalServerError)
        return
    }

    version := uuid.New()
    data := PageData{
        Content: template.HTML(content),
        Version: version.String(),
    }

    if err := tmpl.ExecuteTemplate(w, "contact", data); err != nil {
        log.Println(err)
        http.Error(w, "Error executing template.", http.StatusInternalServerError)
    }
}

func handleHome(w http.ResponseWriter, r *http.Request){

    tmpl, err := ParseTemplates()
    if err != nil {
        log.Printf("Error loading templates: %v\n", err)
        http.Error(w, "Error loading templates.", http.StatusInternalServerError)
        return
    }

    content, err := RenderTemplate(tmpl, "home", nil)
    if err != nil {
        log.Printf("Error rendering contact template: %v\n", err)
        http.Error(w, "Error rendering template.", http.StatusInternalServerError)
        return
    }

    version := uuid.New()
    data := PageData{
        Content: template.HTML(content),
        Version: version.String(),
    }

    if err := tmpl.ExecuteTemplate(w, "home", data); err != nil {
        log.Println(err)
        http.Error(w, "Error executing template.", http.StatusInternalServerError)
    }
}

func handleBlog(w http.ResponseWriter, r *http.Request){
    tmpl, err := ParseTemplates()
    if err != nil {
        log.Printf("Error loading templates: %v\n", err)
        http.Error(w, "Error loading templates.", http.StatusInternalServerError)
        return
    }

    data, err := GetPosts()
    if err != nil{
        log.Printf("Error fetching Posts: %v\n", err)
        http.Error(w, "Error fetching posts.", http.StatusInternalServerError)
    }

    if err := tmpl.ExecuteTemplate(w, "blog", data); err != nil {
        log.Println(err)
        http.Error(w, "Error executing template.", http.StatusInternalServerError)
    }
}

func handleLikeIncrement(w http.ResponseWriter, r *http.Request){
    tmpl, err := ParseTemplates()
    if err != nil {
        log.Printf("Error loading templates: %v\n", err)
        http.Error(w, "Error loading templates.", http.StatusInternalServerError)
        return
    }

    if query := r.URL.Query(); !query.Has("id"){
        http.Error(w, "Post id is required", http.StatusBadRequest)
        return
    }

    idStr := r.URL.Query().Get("id")

    id, err := primitive.ObjectIDFromHex(idStr)
    if err != nil{
        log.Println(err)
        http.Error(w, "Invalid Id", http.StatusBadRequest)
        return
    }

    data, err := IncrementLike(id)
    if err != nil{
        log.Println(err)
        http.Error(w, "Error incrementing likes", http.StatusInternalServerError)
        return
    }

    if err := tmpl.ExecuteTemplate(w, "like-button", data); err != nil{ log.Println(err)
    http.Error(w, "Error executing template.", http.StatusInternalServerError)
}
}

type PostReadData struct{
    Posts []Post
    Post Post
    MarkDown template.HTML
}

func handleReadPost(w http.ResponseWriter, r *http.Request){
    if r.Method != http.MethodGet{
        http.Error(w, "Only get method is accepted", http.StatusMethodNotAllowed)
        return
    }


    if query := r.URL.Query(); !query.Has("id"){
        http.Error(w, "Post id is required", http.StatusBadRequest)
        return
    }

    idStr := r.URL.Query().Get("id")

    posts, err := GetPosts()
    if err != nil{
        log.Println(err)
        http.Error(w, "Error trying to fetch posts", http.StatusInternalServerError)
        return
    }

    post, err := GetPost(idStr)
    if err != nil{
        log.Println(err)
        http.Error(w, "Invalid id provided", http.StatusInternalServerError)
        return
    }

    markDown := template.HTML(MdToHtml([]byte(post.Body)))

    data := PostReadData{
        Posts: RemovePostFromList(posts, idStr),
        Post: post,
        MarkDown: markDown,
    }

    tmpl, err := ParseTemplates()
    if err != nil {
        log.Printf("Error loading templates: %v\n", err)
        http.Error(w, "Error loading templates.", http.StatusInternalServerError)
        return
    }

    if err := tmpl.ExecuteTemplate(w, "read-post", data); err != nil{
        log.Println(err)
        http.Error(w, "Error executing template.", http.StatusInternalServerError)
    }
}

func handleSearchPosts(w http.ResponseWriter, r *http.Request){
    if r.Method != http.MethodGet{
        http.Error(w, "Http method not allowed", http.StatusMethodNotAllowed)
        return
    }

    search := r.URL.Query().Get("search-text")
    postID := r.URL.Query().Get("postID")

    escapedSearch := regexp.QuoteMeta(search)
    pattern := fmt.Sprintf("^%s", escapedSearch)

    objectID, err := primitive.ObjectIDFromHex(postID)

    var posts []Post
    if err != nil{
        log.Println(err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if len(escapedSearch) == 0{
        posts, err = GetPosts()
        posts = RemovePostFromList(posts, postID)
    } else{
        filter := bson.M{
            "title": primitive.Regex{Pattern: pattern, Options: "i"},
            "_id": bson.M{"$ne": objectID},
        }

        posts, err = QueryPosts(filter)
    }

    if err != nil{
        log.Println(err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }

    tmpl, err := ParseTemplates()
    if err != nil {
        log.Printf("Error loading templates: %v\n", err)
        http.Error(w, "Error loading templates.", http.StatusInternalServerError)
        return
    }

    if err := tmpl.ExecuteTemplate(w, "posts-list", posts); err != nil{
        log.Println(err)
        http.Error(w, "Error executing template.", http.StatusInternalServerError)
    }
}
