package main

import (
	"context"
	"errors"
	"log"
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
const users_col string = "users"

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
    Id       primitive.ObjectID `bson:"_id,omitempty"`
    Title    string             `bson:"title,omitempty"`
    Body     string             `bson:"body,omitempty"`
    Date     time.Time          `bson:"date"`
}

func CreatePost(title, body string) (Post, error){
    post := Post{ 
        Title: title,
        Body: body,
        Date: time.Now(),
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

   var posts []Post
   cur, err := collection.Find(ctx, bson.D{})

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

func UpdatePost(id primitive.ObjectID, title string, body string) (Post, error){
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
        },
    }

    _, err = collection.UpdateOne(
        ctx,
        bson.M{"_id": id},
        update,
    )

    post.Title = title
    post.Body = body

    return post, err
}
