package main

import(
    "net/http"
    "log"
    "html/template"
    "github.com/google/uuid"
    "context"
    "time"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

func HandleAdminEndpoints(){
    http.HandleFunc("/admin", handleAdmin)
    http.HandleFunc("/authenticate", handleAuthentication)
}


func handleAdmin(w http.ResponseWriter, r *http.Request){
    if !isAuthenticated(r){
        showLoginForm(w, r)
    } else{
        showDashboard(w, r)
    }
}

func isAuthenticated(r *http.Request) bool{
    cookie, err := r.Cookie("session_token")
    if err != nil {
        return false
    }

    collection := Client.Database("yourdb").Collection("sessions")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    var session struct {
        UserID  primitive.ObjectID `bson:"user_id"`
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

func showDashboard(w http.ResponseWriter, r *http.Request) {
    tmpl, err := ParseTemplates()
    if err != nil {
        log.Printf("Error loading templates: %v\n", err)
        http.Error(w, "Error loading templates.", http.StatusInternalServerError)
        return
    }
    content, err := RenderTemplate(tmpl, "dashboard", nil)
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
    if err := tmpl.ExecuteTemplate(w, "dashboard", data); err != nil {
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

func storeSession(userID primitive.ObjectID, token string, expiresAt time.Time) error {
    collection := Client.Database("yourdb").Collection("sessions")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    _, err := collection.InsertOne(ctx, bson.M{
        "user_id": userID,
        "token":   token,
        "expires": expiresAt,
    })
    return err
}
