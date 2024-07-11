package main

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
    "context"
    "log"
    "time"
    "errors"
)

var Client *mongo.Client

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

    collection := Client.Database("blog").Collection("users")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    _, err = collection.InsertOne(ctx, user)
    return err
}

func AuthenticateUser(username, password string) (*User, error) {
    collection := Client.Database("blog").Collection("users")
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
