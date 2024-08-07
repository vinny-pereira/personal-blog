package repository

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"sync"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var Client *mongo.Client
const db string = "blog"
const posts_col string = "posts"
const portfolio_col string = "portfolio"
const users_col string = "users"

type AppRepository struct{
    posts map[primitive.ObjectID]Post
    lock  *sync.Mutex
}

type PageData struct {
	Content template.HTML
	Version string
}

type PortfolioEntry struct {
    Id         primitive.ObjectID `bson:"_id,omitempty"`
    Title      string             `bson:"title,omitempty"`
    Date       time.Time          `bson:"date"`
    Repo       string             `bson:"repo"`
    Url        string             `bson:"url"`
    CoverImage string             `bson:"coverimage"`
}

func ConnectMongoDB() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    ClientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
    var err error
    Client, err = mongo.Connect(ctx, ClientOptions)
    if err != nil {
        log.Fatal(err)
    }

    err = Client.Ping(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Connected to MongoDB")
}

type User struct {
    Id       primitive.ObjectID `bson:"_id,omitempty"`
    Username string             `bson:"username"`
    Password string             `bson:"password"`
}

func (u *User) HashPassword() error {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u.Password = string(hashedPassword)
    return nil
}

func (u *User) CheckPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
    return err == nil
}

func RegisterUser(username, password string) error {
    user := User{
        Username: username,
        Password: password,
    }

    err := user.HashPassword()
    if err != nil {
        return err
    }

    collection := Client.Database(db).Collection(users_col)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    _, err = collection.InsertOne(ctx, user)
    return err
}

func AuthenticateUser(username, password string) (*User, error) {
    collection := Client.Database(db).Collection(users_col)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    var user User
    err := collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
    if err != nil {
        return nil, err
    }

    if !user.CheckPassword(password) {
        return nil, errors.New("invalid password")
    }

    return &user, nil
}

type Post struct{
    Id         primitive.ObjectID `bson:"_id,omitempty"`
    Title      string             `bson:"title,omitempty"`
    Body       string             `bson:"body,omitempty"`
    Date       time.Time          `bson:"date"`
    Synopsys   string             `bson:"synopsys"`
    Likes      int                `bson:"likes"`
    Comments   int                `bson:"comments"`
    CoverImage string             `bson:"coverimage"`
}

func (p Post) MainFormatDate() string {
    return p.Date.Format(time.DateOnly)
} 

func CreatePost(title, body string, synopsys string, coverImage string) (Post, error){
    post := Post{ 
        Title: title,
        Body: body,
        Synopsys: synopsys,
        Date: time.Now(),
        CoverImage: coverImage,
    }

    collection := Client.Database(db).Collection(posts_col)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    _, err := collection.InsertOne(ctx, post)
    return post, err
}

func GetPosts()([]Post, error){
    collection := Client.Database(db).Collection(posts_col)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    opts := options.Find().SetSort(bson.D{{"date", -1}})

    var posts []Post
    cur, err := collection.Find(ctx, bson.D{}, opts)

    if err != nil{
        return posts, err
    }

    err = cur.All(ctx, &posts)

    return posts, err
}

func GetPost(id string) (Post, error) {
    collection := Client.Database(db).Collection(posts_col)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    objectId, err := primitive.ObjectIDFromHex(id)
    if err != nil{
        return Post{}, err
    }

    var post Post 
    err = collection.FindOne(ctx, bson.M{"_id": objectId}).Decode(&post)

    return post, err
}

func UpdatePost(id primitive.ObjectID, title string, body string, synopsys string, coverImage string) (Post, error){
    collection := Client.Database(db).Collection(posts_col)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    post, err := GetPost(id.Hex())
    if err != nil{
        return post, err
    }

    update := bson.M{
        "$set": bson.M{
            "title": title,
            "body":  body,
            "synopsys": synopsys,
            "coverimage": coverImage,
        },
    }

    _, err = collection.UpdateOne(
        ctx,
        bson.M{"_id": id},
        update,
    )

    post.Title = title
    post.Body = body
    post.Synopsys = synopsys
    post.CoverImage = coverImage

    return post, err
}

func DeletePost(id string) error {
    collection := Client.Database(db).Collection(posts_col)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return fmt.Errorf("invalid ID format: %v", err)
    }

    result, err := collection.DeleteOne(ctx, bson.M{"_id": objectID})
    if err != nil {
        return err 
    }

    if result.DeletedCount == 0 {
        return fmt.Errorf("no post found with ID %s", id)
    }

    return nil
}

func IncrementLike(id primitive.ObjectID) (Post, error){
    collection := Client.Database(db).Collection(posts_col)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    post, err := GetPost(id.Hex())
    if err != nil{
        return post, err
    }

    post.Likes++

    update := bson.M{
        "$set": bson.M{
            "likes": post.Likes,
        },
    }

    _, err = collection.UpdateOne(
        ctx,
        bson.M{"_id": id},
        update,
    )

    return post, err
}

func QueryPosts(filter primitive.M)([]Post, error){
    collection := Client.Database(db).Collection(posts_col)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    opts := options.Find().SetSort(bson.D{{"date", -1}})

    var posts []Post
    cur, err := collection.Find(ctx, filter, opts)

    if err != nil{
        return posts, err
    }
    
    err = cur.All(ctx, &posts)

    if err != nil{
        return posts, err
    }

    return posts, nil
}

func GetPortfolioEntries()([]PortfolioEntry, error){
    collection := Client.Database(db).Collection(portfolio_col)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    opts := options.Find().SetSort(bson.D{{"date", -1}})

    var entries []PortfolioEntry
    cur, err := collection.Find(ctx, bson.D{}, opts)

    if err != nil{
        return entries,err
    }

    err = cur.All(ctx, &entries)

    return entries, err
}

func CreatePortfolioEntry(title string, repo string, url string, coverImage string)(PortfolioEntry, error){
    entry := PortfolioEntry{ 
        Title: title,
        Repo: repo,
        Url: url,
        Date: time.Now(),
        CoverImage: coverImage,
    }

    collection := Client.Database(db).Collection(portfolio_col)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    _, err := collection.InsertOne(ctx, entry)
    return entry, err
}

func UpdateEntry(id primitive.ObjectID, title string, repo string, url string, coverImage string)(PortfolioEntry, error){
    collection := Client.Database(db).Collection(portfolio_col)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    entry, err := GetEntry(id.Hex())
    if err != nil{
        return entry, err
    }

    update := bson.M{
        "$set": bson.M{
            "title": title,
            "repo":  repo,
            "url": url,
            "coverimage": coverImage,
        },
    }

    _, err = collection.UpdateOne(
        ctx,
        bson.M{"_id": id},
        update,
    )

    entry.Title = title
    entry.Repo = repo
    entry.Url = url
    entry.CoverImage = coverImage

    return entry, err
}

func GetEntry(id string)(PortfolioEntry, error){
    collection := Client.Database(db).Collection(portfolio_col)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    objectId, err := primitive.ObjectIDFromHex(id)
    if err != nil{
        return PortfolioEntry{}, err
    }

    var entry PortfolioEntry
    err = collection.FindOne(ctx, bson.M{"_id": objectId}).Decode(&entry)

    return entry, err
}

func DeleteEntry(id string)(error){
    collection := Client.Database(db).Collection(portfolio_col)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return fmt.Errorf("invalid ID format: %v", err)
    }

    result, err := collection.DeleteOne(ctx, bson.M{"_id": objectID})
    if err != nil {
        return err 
    }

    if result.DeletedCount == 0 {
        return fmt.Errorf("no post found with ID %s", id)
    }

    return nil
}
